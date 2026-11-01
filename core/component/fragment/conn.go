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
	"io"
	"net"
	"sync"
	"time"
)

type fragConn struct {
	net.Conn
	opt          Option
	buf          []byte
	mu           sync.Mutex
	fragmentDone bool
	closed       bool
	lastErr      error
	timer        *time.Timer
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

// targetDesc returns a privacy-masked identifier for this connection (SNI + RemoteAddr).
func (fc *fragConn) targetDesc(sni string) string {
	remote := ""
	if fc.Conn != nil && fc.Conn.RemoteAddr() != nil {
		remote = fc.Conn.RemoteAddr().String()
	}
	return TargetIdentifier(remote, sni)
}

// chunkLengths returns slice of byte lengths for a list of chunks.
func chunkLengths(chunks [][]byte) []int {
	res := make([]int, len(chunks))
	for i, c := range chunks {
		res[i] = len(c)
	}
	return res
}

// writeAll writes data in a loop, guaranteeing all bytes are written or returning an error.
func (fc *fragConn) writeAll(data []byte) (int, error) {
	total := 0
	for len(data) > 0 {
		n, err := fc.Conn.Write(data)
		if n > 0 {
			total += n
			data = data[n:]
		}
		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, io.ErrShortWrite
		}
	}
	return total, nil
}

func (fc *fragConn) resetTimer() {
	timeout := fc.opt.GetBufferTimeout()
	if fc.timer == nil {
		fc.timer = time.AfterFunc(timeout, fc.onBufferTimeout)
	} else {
		fc.timer.Reset(timeout)
	}
}

func (fc *fragConn) stopTimer() {
	if fc.timer != nil {
		fc.timer.Stop()
	}
}

func (fc *fragConn) onBufferTimeout() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.fragmentDone || fc.closed || len(fc.buf) == 0 {
		return
	}

	fc.fragmentDone = true
	toWrite := fc.buf
	fc.buf = nil

	RecordBufferTimeout()

	timeout := fc.opt.GetBufferTimeout()
	target := fc.targetDesc("")
	LogWarn("[%s] WARNING: Buffer timeout (%v) reached with %d bytes, flushing as raw data", target, timeout, len(toWrite))

	_, err := fc.writeAll(toWrite)
	if err != nil {
		fc.lastErr = err
		RecordFragmentationError(string(fc.opt.GetMode()), err)
		LogError("[%s] ERROR: Failed to flush buffer on timeout (%d bytes): %v", target, len(toWrite), err)
	}
}

func (fc *fragConn) Read(p []byte) (int, error) {
	fc.mu.Lock()
	err := fc.lastErr
	fc.mu.Unlock()
	if err != nil {
		return 0, err
	}
	return fc.Conn.Read(p)
}

func (fc *fragConn) Write(p []byte) (int, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.lastErr != nil {
		return 0, fc.lastErr
	}

	// If fragmentation was already completed or disabled, pass through transparently
	if fc.fragmentDone || !fc.opt.Enabled {
		return fc.Conn.Write(p)
	}

	// Fast-path / safety check:
	// A TLS ClientHello record begins with:
	// byte 0: ContentType = 0x16 (Handshake)
	// byte 1-2: ProtocolVersion = 0x03 0x01..0x03
	if len(fc.buf) == 0 {
		if len(p) == 0 {
			return 0, nil
		}
		// If the very first byte is not 0x16, this cannot be a TLS Handshake record.
		if p[0] != 0x16 {
			fc.fragmentDone = true
			fc.stopTimer()
			RecordNonTLSBypass()
			LogDebug("[%s] Non-TLS stream detected (prefix: 0x%02x), bypassing fragmentation", fc.targetDesc(""), p[0])
			return fc.Conn.Write(p)
		}
		// If at least 2 bytes are provided and byte 1 is not 0x03, this is not TLS.
		if len(p) >= 2 && p[1] != 0x03 {
			fc.fragmentDone = true
			fc.stopTimer()
			RecordNonTLSBypass()
			LogDebug("[%s] Non-TLS stream detected (byte 1: 0x%02x != 0x03), bypassing fragmentation", fc.targetDesc(""), p[1])
			return fc.Conn.Write(p)
		}
		// If exactly 1 byte was written (0x16), buffer it and wait for subsequent bytes.
	}

	prevLen := len(fc.buf)
	fc.buf = append(fc.buf, p...)

	// If only 1 byte (0x16) is buffered so far, start timer and wait for the second byte
	if len(fc.buf) < 2 {
		fc.resetTimer()
		return len(p), nil
	}

	// If the second byte is not 0x03, it's not a TLS Handshake record -> flush immediately as-is
	if fc.buf[1] != 0x03 {
		fc.stopTimer()
		fc.fragmentDone = true
		toWrite := fc.buf
		fc.buf = nil
		RecordNonTLSBypass()
		LogDebug("[%s] Non-TLS stream detected (byte 1: 0x%02x != 0x03), flushing %d bytes as raw stream", fc.targetDesc(""), toWrite[1], len(toWrite))
		nSent, err := fc.writeAll(toWrite)
		n := nSent - prevLen
		if n < 0 {
			n = 0
		}
		if n > len(p) {
			n = len(p)
		}
		if err != nil {
			fc.lastErr = err
			LogError("[%s] ERROR: Failed to write raw stream: %v", fc.targetDesc(""), err)
		}
		return n, err
	}

	// False-positive protection on non-TLS streams:
	// As soon as at least 6 bytes are buffered, verify that byte 5 is 0x01 (ClientHello)
	// and record length (bytes 3-4) does not exceed 16384 (RFC 5246 max plaintext fragment).
	// If it doesn't match, this is not a ClientHello; flush immediately as-is and mark done.
	if len(fc.buf) >= 6 {
		recLen := int(binary.BigEndian.Uint16(fc.buf[3:5]))
		if fc.buf[5] != 0x01 || recLen > 16384 || recLen < 4 {
			fc.stopTimer()
			fc.fragmentDone = true
			toWrite := fc.buf
			fc.buf = nil
			RecordNonTLSBypass()
			LogDebug("[%s] Non-ClientHello stream detected (byte 5: 0x%02x, recLen: %d), flushing %d bytes as raw stream", fc.targetDesc(""), toWrite[5], recLen, len(toWrite))
			nSent, err := fc.writeAll(toWrite)
			n := nSent - prevLen
			if n < 0 {
				n = 0
			}
			if n > len(p) {
				n = len(p)
			}
			if err != nil {
				fc.lastErr = err
				LogError("[%s] ERROR: Failed to write raw stream: %v", fc.targetDesc(""), err)
			}
			return n, err
		}
	}

	// We need at least 5 bytes to determine total TLS record length
	if len(fc.buf) < 5 {
		fc.resetTimer()
		return len(p), nil
	}

	recLen := int(binary.BigEndian.Uint16(fc.buf[3:5]))
	totalExpected := 5 + recLen

	// If we haven't received the full ClientHello record yet, keep buffering and reset timeout
	if len(fc.buf) < totalExpected {
		fc.resetTimer()
		return len(p), nil
	}

	// Full ClientHello record accumulated!
	fc.stopTimer()
	fc.fragmentDone = true

	// Ensure TCP_NODELAY is active so the OS stack won't coalesce fragments
	setTCPNoDelay(fc.Conn)

	helloRecord := fc.buf[:totalExpected]
	extraBytes := fc.buf[totalExpected:]
	fc.buf = nil

	sni := ExtractSNI(helloRecord)
	target := fc.targetDesc(sni)
	mode := fc.opt.GetMode()

	RecordClientHelloDetected(string(mode))
	LogInfo("[%s] TLS ClientHello detected (size: %d bytes, mode: %s)", target, len(helloRecord), mode)

	// Fragment and send the ClientHello
	startTime := time.Now()
	nFrag, err := fc.sendFragmented(helloRecord, target, sni)
	elapsed := time.Since(startTime)
	if err != nil {
		RecordFragmentationError(string(mode), err)
		n := nFrag - prevLen
		if n < 0 {
			n = 0
		}
		if n > len(p) {
			n = len(p)
		}
		fc.lastErr = err
		return n, err
	}

	// If there were any extra bytes buffered beyond the ClientHello record, send them directly
	if len(extraBytes) > 0 {
		LogDebug("[%s] Sending %d extra buffered bytes directly", target, len(extraBytes))
		nExtra, err := fc.writeAll(extraBytes)
		if err != nil {
			RecordFragmentationError(string(mode), err)
			n := (nFrag + nExtra) - prevLen
			if n < 0 {
				n = 0
			}
			if n > len(p) {
				n = len(p)
			}
			fc.lastErr = err
			LogError("[%s] ERROR: Failed to write extra buffered bytes: %v", target, err)
			return n, err
		}
	}

	RecordFragmentationSuccess(string(mode))
	_ = elapsed
	return len(p), nil
}

func (fc *fragConn) Close() error {
	fc.mu.Lock()
	fc.stopTimer()
	var flushErr error
	if !fc.fragmentDone && len(fc.buf) > 0 {
		fc.fragmentDone = true
		toWrite := fc.buf
		fc.buf = nil
		_, flushErr = fc.writeAll(toWrite)
	}
	fc.closed = true
	lastErr := fc.lastErr
	fc.mu.Unlock()

	closeErr := fc.Conn.Close()
	if flushErr != nil {
		return flushErr
	}
	if lastErr != nil {
		return lastErr
	}
	return closeErr
}

func (fc *fragConn) sendFragmented(record []byte, target string, sni string) (int, error) {
	// Fallback check: if record is corrupted or too short, write as-is
	if len(record) < 5 || record[0] != 0x16 || record[1] != 0x03 {
		return fc.writeAll(record)
	}

	mode := fc.opt.GetMode()
	isRecordFrag := fc.opt.IsRecordFragment()
	startTime := time.Now()

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
		LogDebug("[%s] Mode: RECORD_FRAGMENT, TLS records created: %d (record lengths: %v)", target, len(recordsToSend), chunkLengths(recordsToSend))

	case ModeBoth:
		// BOTH TLS record fragmentation AND TCP segmentation:
		// Split each TLS record into multiple TCP chunks using configured length range
		for _, rec := range recordsToSend {
			chunks = append(chunks, splitRandom(rec, &fc.opt)...)
		}
		LogDebug("[%s] Mode: BOTH, TLS records: %d, TCP chunks: %d (chunk lengths: %v)", target, len(recordsToSend), len(chunks), chunkLengths(chunks))

	case ModeRandom:
		// Conflict resolution: if record-fragment is also true, split into records, then each into chunks
		if isRecordFrag {
			for _, rec := range recordsToSend {
				chunks = append(chunks, splitRandom(rec, &fc.opt)...)
			}
			LogDebug("[%s] Mode: RANDOM+RECORD_FRAGMENT, TLS records: %d, TCP chunks: %d (lengths: %v)", target, len(recordsToSend), len(chunks), chunkLengths(chunks))
		} else {
			chunks = splitRandom(record, &fc.opt)
			LogDebug("[%s] Mode: RANDOM, TCP chunks: %d (lengths: %v)", target, len(chunks), chunkLengths(chunks))
		}

	case ModeSNI:
		fallthrough
	default:
		// Conflict resolution: if record-fragment is also true, send records as chunks
		if isRecordFrag {
			chunks = recordsToSend
			LogDebug("[%s] Mode: SNI+RECORD_FRAGMENT, TLS records created: %d (record lengths: %v)", target, len(recordsToSend), chunkLengths(recordsToSend))
		} else {
			cut := -1
			sniOffset, sniLen, ok := FindSNIOffset(record)
			if ok && sniLen > 1 {
				minCut, maxCut := fc.opt.ParseLength(1, sniLen-1)
				if minCut < 1 {
					minCut = 1
				}
				if maxCut >= sniLen {
					maxCut = sniLen - 1
				}
				if minCut > maxCut {
					minCut = 1
					maxCut = sniLen - 1
				}
				cutInDomain := randomBetween(minCut, maxCut)
				cut = sniOffset + cutInDomain
				if cut > 5 && cut < len(record) {
					chunks = [][]byte{record[:cut], record[cut:]}
					LogDebug("[%s] Mode: SNI (server_name: %s), split at boundary offset %d (chunk lengths: %v)", target, MaskDomain(sni), cut, chunkLengths(chunks))
				}
			}
			if len(chunks) == 0 {
				chunks = splitRandom(record, &fc.opt)
				LogDebug("[%s] Mode: SNI fallback to random (SNI not found), TCP chunks: %d (lengths: %v)", target, len(chunks), chunkLengths(chunks))
			}
		}
	}

	// Step 3: Write TCP chunks with inter-packet delays selected separately per segment pair
	minI, maxI := fc.opt.IntervalRange(1, 5)
	totalSent := 0
	for cIdx, chunk := range chunks {
		var pause time.Duration
		if cIdx > 0 {
			pause = fc.opt.RandomInterval(minI, maxI)
			if pause > 0 {
				time.Sleep(pause)
			}
		}
		LogDebug("[%s] TCP write %d/%d (%d bytes, pause: %v)", target, cIdx+1, len(chunks), len(chunk), pause)
		n, err := fc.writeAll(chunk)
		totalSent += n
		if err != nil {
			LogError("[%s] ERROR: Failed to write fragment %d/%d (%d bytes): %v", target, cIdx+1, len(chunks), len(chunk), err)
			return totalSent, err
		}
	}

	elapsed := time.Since(startTime)
	LogInfo("[%s] Fragmentation completed successfully (mode: %s, TCP writes: %d, elapsed: %v)", target, mode, len(chunks), elapsed)

	return totalSent, nil
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
// If SNI is present, cuts within the SNI domain based on opt.ParseLength(1, sniLen-1).
// If SNI is not present or in record_fragment without SNI, splits into N valid TLS records
// sized according to opt.ParseLength(min, max).
func splitRecord(record []byte, opt *Option) [][]byte {
	if len(record) <= 5 {
		return [][]byte{record}
	}

	version := record[1:3]
	hsPayload := record[5:]
	if len(hsPayload) <= 1 {
		return [][]byte{record}
	}

	sniOffset, sniLen, ok := FindSNIOffset(record)
	if ok && sniLen > 1 {
		domainStart := sniOffset - 5
		minCut, maxCut := opt.ParseLength(1, sniLen-1)
		if minCut < 1 {
			minCut = 1
		}
		if maxCut >= sniLen {
			maxCut = sniLen - 1
		}
		if minCut > maxCut {
			minCut = 1
			maxCut = sniLen - 1
		}
		cutInDomain := randomBetween(minCut, maxCut)
		cut := domainStart + cutInDomain
		if cut > 0 && cut < len(hsPayload) {
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
	}

	// Fallback or no-SNI: split hsPayload into N valid TLS records of size [minLen, maxLen]
	minLen, maxLen := opt.ParseLength(10, 30)
	if minLen <= 0 {
		minLen = 1
	}
	if maxLen < minLen {
		maxLen = minLen
	}

	var records [][]byte
	remaining := hsPayload
	for len(remaining) > 0 {
		chunkSize := randomBetween(minLen, maxLen)
		if chunkSize >= len(remaining) {
			chunkSize = len(remaining)
		}
		recPayload := remaining[:chunkSize]
		rec := make([]byte, 5+len(recPayload))
		rec[0] = 0x16
		copy(rec[1:3], version)
		binary.BigEndian.PutUint16(rec[3:5], uint16(len(recPayload)))
		copy(rec[5:], recPayload)
		records = append(records, rec)
		remaining = remaining[chunkSize:]
	}

	return records
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
	setTCPNoDelayFallback(c)
}
