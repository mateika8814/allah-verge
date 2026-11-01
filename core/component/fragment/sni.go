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
	"strings"
)

// FindSNIOffset parses a TLS ClientHello record to find the exact byte offset
// and length of the server_name host in the record buffer.
// It returns (offset, length, true) if found, or (0, 0, false) if not found or on parse failure.
func FindSNIOffset(b []byte) (offset int, length int, ok bool) {
	// A minimum TLS ClientHello record is at least 44 bytes
	if len(b) < 44 {
		return 0, 0, false
	}

	// Must be a TLS Handshake record (0x16) with major version 0x03
	if b[0] != 0x16 || b[1] != 0x03 {
		return 0, 0, false
	}

	recLen := int(binary.BigEndian.Uint16(b[3:5]))
	if len(b) < 5+recLen {
		return 0, 0, false
	}

	// Handshake Message starts at offset 5
	// Handshake type: 0x01 = ClientHello
	if b[5] != 0x01 {
		return 0, 0, false
	}

	// Handshake length (3 bytes)
	hsLen := int(b[6])<<16 | int(b[7])<<8 | int(b[8])
	if hsLen < 38 || len(b) < 5+4+hsLen {
		return 0, 0, false
	}

	// ClientVersion (2 bytes) + Random (32 bytes) = 34 bytes
	// Offset of SessionID length: 5 + 4 + 34 = 43
	p := 43
	if len(b) < p+1 {
		return 0, 0, false
	}
	sidLen := int(b[p])
	p += 1 + sidLen

	// CipherSuites: 2 bytes length + data
	if len(b) < p+2 {
		return 0, 0, false
	}
	csLen := int(binary.BigEndian.Uint16(b[p : p+2]))
	p += 2 + csLen

	// CompressionMethods: 1 byte length + data
	if len(b) < p+1 {
		return 0, 0, false
	}
	cmLen := int(b[p])
	p += 1 + cmLen

	// Extensions: 2 bytes length + extensions
	if len(b) < p+2 {
		return 0, 0, false
	}
	extsLen := int(binary.BigEndian.Uint16(b[p : p+2]))
	p += 2
	endExts := p + extsLen
	if endExts > len(b) {
		endExts = len(b)
	}

	// Traverse extensions
	for p+4 <= endExts {
		extType := binary.BigEndian.Uint16(b[p : p+2])
		extLen := int(binary.BigEndian.Uint16(b[p+2 : p+4]))
		p += 4

		if p+extLen > endExts {
			return 0, 0, false
		}

		if extType == 0x0000 { // Server Name Indication (SNI)
			sniData := b[p : p+extLen]
			// SNI list length (2 bytes) + name type (1 byte) + name length (2 bytes)
			if len(sniData) >= 5 && sniData[2] == 0x00 { // 0x00 = host_name
				hostLen := int(binary.BigEndian.Uint16(sniData[3:5]))
				if hostLen > 0 && len(sniData) >= 5+hostLen {
					hostOffset := p + 5
					return hostOffset, hostLen, true
				}
			}
		}

		p += extLen
	}

	return 0, 0, false
}

// ExtractSNI extracts the server_name (SNI) string from a TLS ClientHello record if present.
func ExtractSNI(b []byte) string {
	offset, length, ok := FindSNIOffset(b)
	if ok && offset >= 0 && offset+length <= len(b) {
		return string(b[offset : offset+length])
	}
	return ""
}

// MaskDomain masks sensitive middle parts of a domain name to protect user privacy in logs.
// e.g. "youtube.com" -> "y***e.com", "api.github.com" -> "***.g***b.com".
func MaskDomain(d string) string {
	if len(d) == 0 {
		return ""
	}
	parts := strings.Split(d, ".")
	if len(parts) >= 2 {
		mainIdx := len(parts) - 2
		main := parts[mainIdx]
		if len(main) <= 2 {
			parts[mainIdx] = "**"
		} else {
			parts[mainIdx] = string(main[0]) + "***" + string(main[len(main)-1])
		}
		for i := 0; i < mainIdx; i++ {
			parts[i] = "***"
		}
		return strings.Join(parts, ".")
	}
	if len(d) <= 2 {
		return "**"
	}
	return string(d[0]) + "***" + string(d[len(d)-1])
}
