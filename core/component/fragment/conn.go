// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fragment

import (
	"encoding/binary"
	"net"
	"sync"
	"syscall"
	"time"
)

type fragConn struct {
	net.Conn
	opt          Option
	buf          []byte
	mu           sync.Mutex
	fragmentDone bool
}

// NewConn wraps a net.Conn with fragmentation capabilities according to Option.
// If opt is nil or !opt.Enabled, it returns the original connection unchanged.
func NewConn(c net.Conn, opt *Option) net.Conn {
	if c == nil || opt == nil || !opt.Enabled {
		return c
	}
	return &fragConn{
		Conn: c,
		opt:  *opt,
	}
}

func (fc *fragConn) Write(p []byte) (int, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// If fragmentation was already completed or disabled, pass through transparently
	if fc.fragmentDone || !fc.opt.Enabled {
		return fc.Conn.Write(p)
	}

	// Fast-path / safety check:
	// If the initial bytes are not a TLS Handshake (0x16 0x03), this is not a TLS ClientHello.
	// Mark as done immediately and forward as-is.
	if len(fc.buf) == 0 {
		if len(p) < 2 || p[0] != 0x16 || p[1] != 0x03 {
			fc.fragmentDone = true
			return fc.Conn.Write(p)
		}
	}

	// Buffer incoming chunks (e.g. uTLS / Reality writing in small pieces)
	fc.buf = append(fc.buf, p...)

	// We need at least 5 bytes to determine the TLS record length
	if len(fc.buf) < 5 {
		return len(p), nil
	}

	recLen := int(binary.BigEndian.Uint16(fc.buf[3:5]))
	totalExpected := 5 + recLen

	// If we haven't received the full ClientHello record yet, keep buffering
	if len(fc.buf) < totalExpected {
		return len(p), nil
	}

	// Full ClientHello record accumulated!
	// Mark done so any subsequent Write calls pass through with zero overhead.
	fc.fragmentDone = true

	// Ensure TCP_NODELAY is active so the OS stack won't coalesce fragments
	setTCPNoDelay(fc.Conn)

	helloRecord := fc.buf[:totalExpected]
	extraBytes := fc.buf[totalExpected:]

	// Fragment and send the ClientHello
	if err := fc.sendFragmented(helloRecord); err != nil {
		return len(p), err
	}

	// If there were any extra bytes buffered beyond the ClientHello record, send them directly
	if len(extraBytes) > 0 {
		if _, err := fc.Conn.Write(extraBytes); err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

func (fc *fragConn) sendFragmented(record []byte) error {
	defer func() {
		// Clear buffer to free memory
		fc.buf = nil
	}()

	// Fallback check: if record is corrupted or too short, write as-is
	if len(record) < 5 || record[0] != 0x16 || record[1] != 0x03 {
		_, err := fc.Conn.Write(record)
		return err
	}

	mode := fc.opt.GetMode()
	interval := fc.opt.ParseInterval(1, 5)
	isRecordFrag := fc.opt.IsRecordFragment()

	// Step 1: If Record-Fragment is enabled, convert the single TLS record
	// into multiple valid TLS records.
	var recordsToSend [][]byte
	if isRecordFrag && len(record) > 5 {
		recordsToSend = splitRecord(record, &fc.opt)
	} else {
		recordsToSend = [][]byte{record}
	}

	// Step 2: Determine TCP chunks to send
	var chunks [][]byte

	switch mode {
	case ModeRecordFragment:
		// Pure TLS record fragmentation: each TLS record is written as a separate TCP packet
		chunks = recordsToSend

	case ModeBoth:
		// BOTH TLS record fragmentation AND TCP segmentation:
		// Split each TLS record into multiple TCP chunks using configured length range
		for _, rec := range recordsToSend {
			chunks = append(chunks, splitRandom(rec, &fc.opt)...)
		}

	case ModeRandom:
		if isRecordFrag {
			for _, rec := range recordsToSend {
				chunks = append(chunks, splitRandom(rec, &fc.opt)...)
			}
		} else {
			chunks = splitRandom(record, &fc.opt)
		}

	case ModeSNI:
		fallthrough
	default:
		if isRecordFrag {
			chunks = recordsToSend
		} else {
			sniOffset, sniLen, ok := FindSNIOffset(record)
			if ok && sniLen > 0 {
				cut := sniOffset + sniLen/2
				if cut > 5 && cut < len(record) {
					chunks = [][]byte{record[:cut], record[cut:]}
				}
			}
			if len(chunks) == 0 {
				chunks = splitRandom(record, &fc.opt)
			}
		}
	}

	// Step 3: Write TCP chunks with inter-packet delays
	for cIdx, chunk := range chunks {
		if cIdx > 0 && interval > 0 {
			time.Sleep(interval)
		}
		if _, err := fc.Conn.Write(chunk); err != nil {
			return err
		}
	}

	return nil
}

// splitRandom cuts data into chunks of random length within the configured range.
func splitRandom(data []byte, opt *Option) [][]byte {
	minLen, maxLen := opt.ParseLength(10, 30)
	var chunks [][]byte
	remaining := data

	for len(remaining) > 0 {
		cut := randomBetween(minLen, maxLen)
		if cut >= len(remaining) {
			chunks = append(chunks, remaining)
			break
		}
		chunks = append(chunks, remaining[:cut])
		remaining = remaining[cut:]
	}

	return chunks
}

// splitRecord splits a single TLS Handshake record into multiple valid TLS records.
func splitRecord(record []byte, opt *Option) [][]byte {
	if len(record) <= 5 {
		return [][]byte{record}
	}

	version := record[1:3]
	hsPayload := record[5:]
	if len(hsPayload) <= 1 {
		return [][]byte{record}
	}

	// Find split point: cut right in the middle of SNI domain
	var cut int
	sniOffset, sniLen, ok := FindSNIOffset(record)
	if ok && sniLen > 0 {
		// Relative to hsPayload (record header is 5 bytes)
		cut = (sniOffset - 5) + sniLen/2
	}

	if cut <= 0 || cut >= len(hsPayload) {
		minLen, maxLen := opt.ParseLength(10, 30)
		cut = randomBetween(minLen, maxLen)
		if cut >= len(hsPayload) {
			cut = len(hsPayload) / 2
		}
	}
	if cut <= 0 || cut >= len(hsPayload) {
		return [][]byte{record}
	}

	// Record 1
	rec1Payload := hsPayload[:cut]
	rec1 := make([]byte, 5+len(rec1Payload))
	rec1[0] = 0x16
	copy(rec1[1:3], version)
	binary.BigEndian.PutUint16(rec1[3:5], uint16(len(rec1Payload)))
	copy(rec1[5:], rec1Payload)

	// Record 2
	rec2Payload := hsPayload[cut:]
	rec2 := make([]byte, 5+len(rec2Payload))
	rec2[0] = 0x16
	copy(rec2[1:3], version)
	binary.BigEndian.PutUint16(rec2[3:5], uint16(len(rec2Payload)))
	copy(rec2[5:], rec2Payload)

	return [][]byte{rec1, rec2}
}

// setTCPNoDelay enables TCP_NODELAY on the underlying connection to prevent OS packet coalescing.
func setTCPNoDelay(c net.Conn) {
	type noDelayer interface {
		SetNoDelay(bool) error
	}
	if nd, ok := c.(noDelayer); ok {
		_ = nd.SetNoDelay(true)
		return
	}
	if sc, ok := c.(syscall.Conn); ok {
		if rc, err := sc.SyscallConn(); err == nil {
			_ = rc.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
			})
		}
	}
}
