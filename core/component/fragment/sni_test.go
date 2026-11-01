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
	"testing"
)

// Helper to construct a synthetic but RFC-compliant TLS ClientHello
func makeClientHello(sni string, includeSNI bool) []byte {
	var exts []byte
	if includeSNI {
		sniBytes := []byte(sni)
		sniExt := []byte{0x00, 0x00} // extension type 0 = server_name
		sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)+5))
		sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)+3))
		sniExt = append(sniExt, 0x00) // name_type 0 = host_name
		sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)))
		sniExt = append(sniExt, sniBytes...)
		exts = sniExt
	} else {
		// ALPN or dummy extension (e.g. type 0x0010)
		dummy := []byte{0x00, 0x10, 0x00, 0x02, 0x01, 0x02}
		exts = dummy
	}

	hsPayload := []byte{0x03, 0x03}                       // ClientVersion TLS 1.2
	hsPayload = append(hsPayload, make([]byte, 32)...)    // Random (32 bytes)
	hsPayload = append(hsPayload, 0x00)                   // SessionID length 0
	hsPayload = append(hsPayload, 0x00, 0x02, 0x00, 0x9c) // CipherSuites: 1 suite (0x009c)
	hsPayload = append(hsPayload, 0x01, 0x00)             // CompressionMethods: 1 method (null)
	hsPayload = binary.BigEndian.AppendUint16(hsPayload, uint16(len(exts)))
	hsPayload = append(hsPayload, exts...)

	hs := []byte{0x01} // HandshakeType ClientHello
	hsLen := len(hsPayload)
	hs = append(hs, byte(hsLen>>16), byte(hsLen>>8), byte(hsLen))
	hs = append(hs, hsPayload...)

	rec := []byte{0x16, 0x03, 0x01} // Record header: Handshake, TLS 1.0 legacy
	rec = binary.BigEndian.AppendUint16(rec, uint16(len(hs)))
	rec = append(rec, hs...)
	return rec
}

func TestFindSNIOffset_Valid(t *testing.T) {
	testDomains := []string{
		"example.com",
		"subdomain.target.org",
		"a.b.c.d.ru",
		"very-long-domain-name-for-testing-purposes-12345.net",
	}

	for _, domain := range testDomains {
		pkt := makeClientHello(domain, true)
		offset, length, ok := FindSNIOffset(pkt)
		if !ok {
			t.Fatalf("Expected SNI to be found for domain %s", domain)
		}
		if length != len(domain) {
			t.Fatalf("Expected length %d, got %d", len(domain), length)
		}
		extracted := string(pkt[offset : offset+length])
		if extracted != domain {
			t.Fatalf("Extracted domain %s != expected %s", extracted, domain)
		}
	}
}

func TestFindSNIOffset_WithoutSNI(t *testing.T) {
	pkt := makeClientHello("", false)
	_, _, ok := FindSNIOffset(pkt)
	if ok {
		t.Fatalf("Expected FindSNIOffset to return ok=false when SNI is absent")
	}
}

func TestFindSNIOffset_Truncated(t *testing.T) {
	pkt := makeClientHello("example.com", true)

	// Test cutting at every single byte boundary
	for i := 0; i < len(pkt)-1; i++ {
		truncated := pkt[:i]
		_, _, ok := FindSNIOffset(truncated)
		if ok {
			t.Fatalf("Expected ok=false for truncated packet at len %d (original %d)", i, len(pkt))
		}
	}
}

func TestFindSNIOffset_Garbage(t *testing.T) {
	garbageCases := [][]byte{
		nil,
		{},
		{0x00, 0x01, 0x02},
		[]byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
		[]byte("SSH-2.0-OpenSSH_8.9p1\r\n"),
		make([]byte, 100), // all zeros
		{0x16, 0x03, 0x01, 0xFF, 0xFF}, // fake record header with truncated payload
		{0x16, 0x03, 0x01, 0x00, 0x05, 0x02, 0x00, 0x00, 0x01, 0x00}, // Handshake type 2 (ServerHello, not ClientHello)
	}

	for idx, g := range garbageCases {
		_, _, ok := FindSNIOffset(g)
		if ok {
			t.Fatalf("Expected ok=false on garbage case %d", idx)
		}
	}
}
