// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fragment

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestFragConn_NetPipe_SNI(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "my-secret-domain.org"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:  true,
		Mode:     ModeSNI,
		Interval: "1-2",
	}

	fc := NewConn(clientConn, opt)

	receivedChunks := make(chan []byte, 10)
	done := make(chan struct{})

	// Server reading in background
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedChunks <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	// Write ClientHello from client
	n, err := fc.Write(pkt)
	if err != nil || n != len(pkt) {
		t.Fatalf("Write error: n=%d, err=%v", n, err)
	}

	// Write subsequent application data (simulating second and third writes)
	data2 := []byte("SECOND_WRITE_DATA")
	data3 := []byte("THIRD_WRITE_DATA")
	fc.Write(data2)
	fc.Write(data3)

	// Close client to finish reader
	clientConn.Close()
	<-done
	close(receivedChunks)

	var allChunks [][]byte
	var fullReceived []byte
	for c := range receivedChunks {
		allChunks = append(allChunks, c)
		fullReceived = append(fullReceived, c...)
	}

	// Check 1: Full received payload must match original ClientHello + data2 + data3 byte-for-byte
	expectedFull := append(append(pkt, data2...), data3...)
	if !bytes.Equal(fullReceived, expectedFull) {
		t.Fatalf("Reassembled data on receiving end does not match original byte-for-byte!")
	}

	// Check 2: Chunks from ClientHello
	// In SNI mode, ClientHello should be split into 2 chunks
	// Neither chunk should contain the intact SNI string
	if len(allChunks) < 2 {
		t.Fatalf("Expected at least 2 chunks, got %d", len(allChunks))
	}
	chunk1 := allChunks[0]
	chunk2 := allChunks[1]

	if bytes.Contains(chunk1, []byte(sni)) || bytes.Contains(chunk2, []byte(sni)) {
		t.Fatalf("DPI bypass failed: SNI domain found intact in one of the segments!")
	}

	// Check 3: Second and third writes were sent intact
	if !bytes.Contains(fullReceived, data2) || !bytes.Contains(fullReceived, data3) {
		t.Fatalf("Subsequent writes corrupted!")
	}
}

func TestFragConn_NetPipe_Random(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "random-test.net"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:  true,
		Mode:     ModeRandom,
		Length:   "15-25",
		Interval: "1",
	}

	fc := NewConn(clientConn, opt)

	receivedChunks := make(chan []byte, 50)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedChunks <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	fc.Write(pkt)
	clientConn.Close()
	<-done
	close(receivedChunks)

	var fullReceived []byte
	var chunks [][]byte
	for c := range receivedChunks {
		chunks = append(chunks, c)
		fullReceived = append(fullReceived, c...)
	}

	// Check byte-for-byte equality
	if !bytes.Equal(fullReceived, pkt) {
		t.Fatalf("Data mismatch in random mode!")
	}

	// Each chunk (except possibly the last) should be within 15-25 bytes
	for i, c := range chunks[:len(chunks)-1] {
		if len(c) < 15 || len(c) > 25 {
			t.Fatalf("Chunk %d length %d not in range 15-25", i, len(c))
		}
	}
}

func TestFragConn_RecordFragment(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "record-frag.example.com"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:        true,
		Mode:           ModeSNI,
		RecordFragment: true,
		Interval:       "1",
	}

	fc := NewConn(clientConn, opt)

	receivedRecords := make(chan []byte, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedRecords <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	fc.Write(pkt)
	clientConn.Close()
	<-done
	close(receivedRecords)

	var records [][]byte
	for r := range receivedRecords {
		records = append(records, r)
	}

	// Should have received 2 valid TLS records
	if len(records) != 2 {
		t.Fatalf("Expected 2 TLS records, got %d", len(records))
	}

	var totalHandshakePayload []byte
	for i, rec := range records {
		if len(rec) < 5 || rec[0] != 0x16 || rec[1] != 0x03 {
			t.Fatalf("Record %d is not a valid TLS handshake record header", i)
		}
		recLen := int(binary.BigEndian.Uint16(rec[3:5]))
		if len(rec) != 5+recLen {
			t.Fatalf("Record %d declared len %d != actual len %d", i, recLen, len(rec)-5)
		}
		totalHandshakePayload = append(totalHandshakePayload, rec[5:]...)
	}

	// Handshake payload (ClientHello) must match original handshake payload byte-for-byte
	originalHandshakePayload := pkt[5:]
	if !bytes.Equal(totalHandshakePayload, originalHandshakePayload) {
		t.Fatalf("Reassembled handshake payload does not match original!")
	}
}

func TestFragConn_Buffering_UTLS_Reality(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "utls-fragment.org"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:  true,
		Mode:     ModeSNI,
		Interval: "1",
	}

	fc := NewConn(clientConn, opt)

	received := make(chan []byte, 10)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				received <- chunk
			}
			if err != nil {
				close(received)
				return
			}
		}
	}()

	// Feed data in tiny slices (simulating uTLS / Reality writing in small pieces)
	sliceSizes := []int{2, 3, 5, 10, 20}
	pos := 0
	for _, sz := range sliceSizes {
		if pos+sz > len(pkt) {
			break
		}
		fc.Write(pkt[pos : pos+sz])
		pos += sz
		// Ensure nothing is sent while record is incomplete
		select {
		case <-received:
			t.Fatalf("Data was sent before full ClientHello was assembled!")
		case <-time.After(10 * time.Millisecond):
			// ok, still buffering
		}
	}

	// Write remainder
	fc.Write(pkt[pos:])

	// Now full packet should arrive
	var full []byte
	timer := time.After(500 * time.Millisecond)
	loop:
	for {
		select {
		case chunk, ok := <-received:
			if !ok {
				break loop
			}
			full = append(full, chunk...)
			if len(full) >= len(pkt) {
				clientConn.Close()
			}
		case <-timer:
			t.Fatalf("Timed out waiting for reassembled ClientHello")
		}
	}

	if !bytes.Equal(full, pkt) {
		t.Fatalf("Buffered ClientHello mismatch!")
	}
}

func TestFragConn_NonTLS_Passthrough(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	opt := &Option{
		Enabled:  true,
		Mode:     ModeSNI,
		Interval: "1",
	}

	fc := NewConn(clientConn, opt)

	httpData := []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")

	go func() {
		fc.Write(httpData)
		clientConn.Close()
	}()

	readAll, err := io.ReadAll(serverConn)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	if !bytes.Equal(readAll, httpData) {
		t.Fatalf("Non-TLS data was modified or corrupted!")
	}
}

func TestFragConn_Disabled_Passthrough(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	opt := &Option{
		Enabled: false,
	}

	fc := NewConn(clientConn, opt)
	// Should return the exact same conn when disabled
	if fc != clientConn {
		t.Fatalf("Expected NewConn to return raw net.Conn when disabled")
	}
}

func TestFragConn_ModeRecordFragment(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "record-frag-mode.example.com"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:  true,
		Mode:     ModeRecordFragment,
		Interval: "1",
	}

	fc := NewConn(clientConn, opt)

	receivedRecords := make(chan []byte, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedRecords <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	fc.Write(pkt)
	clientConn.Close()
	<-done
	close(receivedRecords)

	var records [][]byte
	for r := range receivedRecords {
		records = append(records, r)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 TLS records for ModeRecordFragment, got %d", len(records))
	}

	// Verify both are valid TLS records
	for i, rec := range records {
		if len(rec) < 5 || rec[0] != 0x16 || rec[1] != 0x03 {
			t.Fatalf("Record %d is not a valid TLS handshake record", i)
		}
	}
}

func TestFragConn_ModeBoth(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "both-mode.example.com"
	pkt := makeClientHello(sni, true)

	opt := &Option{
		Enabled:  true,
		Mode:     ModeBoth,
		Length:   "20-40",
		Interval: "1",
	}

	fc := NewConn(clientConn, opt)

	receivedChunks := make(chan []byte, 20)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedChunks <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	fc.Write(pkt)
	clientConn.Close()
	<-done
	close(receivedChunks)

	var chunks [][]byte
	for c := range receivedChunks {
		chunks = append(chunks, c)
	}

	// In Both mode, we expect multiple TCP chunks (> 2) combining record splitting and length slicing
	if len(chunks) < 3 {
		t.Fatalf("Expected at least 3 chunks in Both mode, got %d", len(chunks))
	}

	// Verify the stream contains 2 TLS record headers (0x16 0x03)
	full := bytes.Join(chunks, nil)
	tlsHeaderCount := 0
	for i := 0; i < len(full)-2; i++ {
		if full[i] == 0x16 && full[i+1] == 0x03 {
			tlsHeaderCount++
		}
	}
	if tlsHeaderCount < 2 {
		t.Fatalf("Expected at least 2 TLS record headers in Both mode, found %d", tlsHeaderCount)
	}
}

func TestFragConn_UnknownMode(t *testing.T) {
	opt := &Option{
		Enabled: true,
		Mode:    "unsupported_mode",
	}

	mode := opt.GetMode()
	if mode != ModeSNI {
		t.Fatalf("Expected fallback to ModeSNI for unknown mode, got %v", mode)
	}
}

func TestFragConn_NonTLS_StartingWith1603(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	opt := &Option{
		Enabled:  true,
		Mode:     ModeSNI,
		Interval: "1-2",
	}

	fc := NewConn(clientConn, opt)

	received := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 1024)
		n, _ := serverConn.Read(buf)
		res := make([]byte, n)
		copy(res, buf[:n])
		received <- res
	}()

	// Fake non-TLS payload: starts with 0x16 0x03, but byte 5 is 0x99 (!= 0x01) and recLen=0xffff > 16384
	nonTLSPkt := []byte{0x16, 0x03, 0xff, 0xff, 0x00, 0x99, 0xde, 0xad, 0xbe, 0xef}
	n, err := fc.Write(nonTLSPkt)
	if err != nil || n != len(nonTLSPkt) {
		t.Fatalf("Expected non-TLS write to succeed immediately, n=%d err=%v", n, err)
	}

	select {
	case got := <-received:
		if !bytes.Equal(got, nonTLSPkt) {
			t.Fatalf("Expected payload to pass through unchanged, got %x, expected %x", got, nonTLSPkt)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Non-TLS payload starting with 16 03 timed out / blocked")
	}
}

func TestFragConn_BufferingTimeout(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	opt := &Option{
		Enabled:       true,
		Mode:          ModeSNI,
		BufferTimeout: 50 * time.Millisecond,
	}

	fc := NewConn(clientConn, opt)

	received := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 1024)
		n, _ := serverConn.Read(buf)
		res := make([]byte, n)
		copy(res, buf[:n])
		received <- res
	}()

	// Partial write: 5 bytes of header claiming length 0x50, but no further bytes ever sent
	partial := []byte{0x16, 0x03, 0x01, 0x00, 0x50}
	n, err := fc.Write(partial)
	if err != nil || n != len(partial) {
		t.Fatalf("Initial write failed: n=%d err=%v", n, err)
	}

	select {
	case got := <-received:
		if !bytes.Equal(got, partial) {
			t.Fatalf("Expected buffered data to flush after timeout, got %x, expected %x", got, partial)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Buffering timeout failed to flush accumulated data")
	}
}

func TestFragConn_CloseWithBufferedData(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	opt := &Option{
		Enabled:       true,
		Mode:          ModeSNI,
		BufferTimeout: 5 * time.Second,
	}

	fc := NewConn(clientConn, opt)

	received := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 1024)
		n, _ := serverConn.Read(buf)
		res := make([]byte, n)
		copy(res, buf[:n])
		received <- res
	}()

	partial := []byte{0x16, 0x03, 0x01, 0x00}
	_, _ = fc.Write(partial)

	// Close immediately with non-empty buffer
	err := fc.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	select {
	case got := <-received:
		if !bytes.Equal(got, partial) {
			t.Fatalf("Expected buffered data to be flushed on Close, got %x, expected %x", got, partial)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Close failed to deliver buffered data")
	}
}

func TestFragConn_Conflict_RandomWithRecordFragment(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "conflict-random-rf.example.com"
	pkt := makeClientHello(sni, true)

	// Conflict configuration: Mode is Random, but RecordFragment is also set to true
	opt := &Option{
		Enabled:        true,
		Mode:           ModeRandom,
		Length:         "10-20",
		Interval:       "1-2",
		RecordFragment: true,
	}

	fc := NewConn(clientConn, opt)

	receivedChunks := make(chan []byte, 20)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedChunks <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	n, err := fc.Write(pkt)
	if err != nil || n != len(pkt) {
		t.Fatalf("Write failed: n=%d err=%v", n, err)
	}
	clientConn.Close()
	<-done
	close(receivedChunks)

	var chunks [][]byte
	for c := range receivedChunks {
		chunks = append(chunks, c)
	}

	// Should produce multiple chunks (> 2) combining record splitting and random chunks
	if len(chunks) < 3 {
		t.Fatalf("Expected at least 3 chunks in Random+RecordFragment mode, got %d", len(chunks))
	}

	full := bytes.Join(chunks, nil)
	tlsHeaderCount := 0
	for i := 0; i < len(full)-2; i++ {
		if full[i] == 0x16 && full[i+1] == 0x03 {
			tlsHeaderCount++
		}
	}
	if tlsHeaderCount < 2 {
		t.Fatalf("Expected at least 2 TLS record headers when RecordFragment=true, found %d", tlsHeaderCount)
	}
}

func TestFragConn_Conflict_SNIWithRecordFragment(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sni := "conflict-sni-rf.example.com"
	pkt := makeClientHello(sni, true)

	// Conflict configuration: Mode is SNI, but RecordFragment is also set to true
	opt := &Option{
		Enabled:        true,
		Mode:           ModeSNI,
		Interval:       "1-2",
		RecordFragment: true,
	}

	fc := NewConn(clientConn, opt)

	receivedChunks := make(chan []byte, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := serverConn.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				receivedChunks <- chunk
			}
			if err != nil {
				return
			}
		}
	}()

	n, err := fc.Write(pkt)
	if err != nil || n != len(pkt) {
		t.Fatalf("Write failed: n=%d err=%v", n, err)
	}
	clientConn.Close()
	<-done
	close(receivedChunks)

	var chunks [][]byte
	for c := range receivedChunks {
		chunks = append(chunks, c)
	}

	// RecordFragment takes precedence: splits into 2 TLS records
	if len(chunks) < 2 {
		t.Fatalf("Expected at least 2 chunks in SNI+RecordFragment mode, got %d", len(chunks))
	}

	full := bytes.Join(chunks, nil)
	tlsHeaderCount := 0
	for i := 0; i < len(full)-2; i++ {
		if full[i] == 0x16 && full[i+1] == 0x03 {
			tlsHeaderCount++
		}
	}
	if tlsHeaderCount < 2 {
		t.Fatalf("Expected at least 2 TLS records when RecordFragment=true, found %d", tlsHeaderCount)
	}
}
