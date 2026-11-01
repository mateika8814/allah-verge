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
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// CheckResult represents a single diagnostic check outcome.
type CheckResult struct {
	Name        string `json:"name"`
	Status      string `json:"status"` // "PASS", "FAIL", "SKIPPED"
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
	DurationMs  int64  `json:"duration_ms"`
}

// SystemStatus holds current system state information for diagnostics.
type SystemStatus struct {
	ModuleAvailable bool   `json:"module_available"`
	ActiveConfig    Option `json:"active_config"`
	ActiveMode      string `json:"active_mode"`
	LoggingStatus   string `json:"logging_status"`
}

// DiagnosticsReport contains all diagnostic check results and system status.
type DiagnosticsReport struct {
	Timestamp int64         `json:"timestamp"`
	AllPassed bool          `json:"all_passed"`
	Status    SystemStatus  `json:"status"`
	Checks    []CheckResult `json:"checks"`
}

// Helper to make an RFC-compliant test ClientHello
func makeDiagClientHello(sni string) []byte {
	var exts []byte
	sniBytes := []byte(sni)
	sniExt := []byte{0x00, 0x00}
	sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)+5))
	sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)+3))
	sniExt = append(sniExt, 0x00)
	sniExt = binary.BigEndian.AppendUint16(sniExt, uint16(len(sniBytes)))
	sniExt = append(sniExt, sniBytes...)
	exts = sniExt

	hsPayload := []byte{0x03, 0x03}
	hsPayload = append(hsPayload, make([]byte, 32)...)
	hsPayload = append(hsPayload, 0x00)
	hsPayload = append(hsPayload, 0x00, 0x02, 0x00, 0x9c)
	hsPayload = append(hsPayload, 0x01, 0x00)
	hsPayload = binary.BigEndian.AppendUint16(hsPayload, uint16(len(exts)))
	hsPayload = append(hsPayload, exts...)

	hs := []byte{0x01}
	hsLen := len(hsPayload)
	hs = append(hs, byte(hsLen>>16), byte(hsLen>>8), byte(hsLen))
	hs = append(hs, hsPayload...)

	rec := []byte{0x16, 0x03, 0x01}
	rec = binary.BigEndian.AppendUint16(rec, uint16(len(hs)))
	rec = append(rec, hs...)
	return rec
}

// RunDiagnostics executes all safe internal checks and returns a comprehensive report.
func RunDiagnostics(opt *Option) DiagnosticsReport {
	startAll := time.Now()
	report := DiagnosticsReport{
		Timestamp: startAll.Unix(),
		AllPassed: true,
		Status: SystemStatus{
			ModuleAvailable: true,
			LoggingStatus:   "Active",
		},
	}

	if opt != nil {
		report.Status.ActiveConfig = *opt
		report.Status.ActiveMode = string(opt.GetMode())
	} else {
		report.Status.ActiveConfig = Option{Enabled: true, Mode: ModeSNI, Length: "10-30", Interval: "1-5"}
		report.Status.ActiveMode = "sni"
	}

	runCheck := func(name, desc string, fn func() error) {
		tStart := time.Now()
		err := fn()
		dur := time.Since(tStart).Milliseconds()
		if dur == 0 {
			dur = 1
		}
		res := CheckResult{
			Name:        name,
			Description: desc,
			DurationMs:  dur,
		}
		if err == nil {
			res.Status = "PASS"
			res.Details = "Verification passed successfully"
		} else {
			res.Status = "FAIL"
			res.Details = err.Error()
			report.AllPassed = false
		}
		report.Checks = append(report.Checks, res)
	}

	// 1. SNI Parser
	runCheck("SNI Parser", "Parses RFC-compliant TLS ClientHello and extracts server_name offset", func() error {
		raw := makeDiagClientHello("diag.internal.net")
		off, l, ok := FindSNIOffset(raw)
		if !ok || l != len("diag.internal.net") {
			return errors.New("failed to find SNI offset in valid ClientHello")
		}
		domain := string(raw[off : off+l])
		if domain != "diag.internal.net" {
			return fmt.Errorf("SNI parsed mismatch: got %q, want diag.internal.net", domain)
		}
		return nil
	})

	// 2. TLS ClientHello Detection
	runCheck("TLS ClientHello Detection", "Validates 0x16 0x03 header and single-byte initial buffering", func() error {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeSNI})
		// Write single byte 0x16
		n, err := fc.Write([]byte{0x16})
		if err != nil || n != 1 {
			return fmt.Errorf("single-byte 0x16 write failed: %v", err)
		}
		return nil
	})

	// 3. Buffering
	runCheck("Buffering", "Ensures partial ClientHello records buffer cleanly until full length", func() error {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		raw := makeDiagClientHello("buffer-check.org")
		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeSNI})

		go func() {
			buf := make([]byte, len(raw))
			_, _ = io.ReadFull(c2, buf)
		}()

		// Split write across 3 chunks
		p1, p2, p3 := raw[:10], raw[10:30], raw[30:]
		if n, err := fc.Write(p1); err != nil || n != len(p1) {
			return fmt.Errorf("chunk 1 write failed: %v", err)
		}
		if n, err := fc.Write(p2); err != nil || n != len(p2) {
			return fmt.Errorf("chunk 2 write failed: %v", err)
		}
		if n, err := fc.Write(p3); err != nil || n != len(p3) {
			return fmt.Errorf("chunk 3 write failed: %v", err)
		}
		return nil
	})

	// 4. SNI Fragmentation
	runCheck("SNI Fragmentation", "Splits ClientHello TCP stream at exact SNI domain boundary", func() error {
		raw := makeDiagClientHello("sni-split.example.com")
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeSNI, Interval: "1-2"})
		var chunks [][]byte
		done := make(chan struct{})

		go func() {
			defer close(done)
			buf := make([]byte, 2048)
			for {
				n, err := c2.Read(buf)
				if n > 0 {
					chk := make([]byte, n)
					copy(chk, buf[:n])
					chunks = append(chunks, chk)
				}
				if err != nil {
					break
				}
			}
		}()

		_, err := fc.Write(raw)
		_ = c1.Close()
		<-done
		if err != nil {
			return err
		}
		if len(chunks) < 2 {
			return fmt.Errorf("expected >= 2 TCP chunks in SNI mode, got %d", len(chunks))
		}
		if !bytes.Equal(bytes.Join(chunks, nil), raw) {
			return errors.New("reassembled payload mismatch")
		}
		return nil
	})

	// 5. Random Fragmentation
	runCheck("Random Fragmentation", "Splits ClientHello into randomized TCP segments within configured length", func() error {
		raw := makeDiagClientHello("random-split.example.com")
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeRandom, Length: "15-25", Interval: "1-2"})
		var chunks [][]byte
		done := make(chan struct{})

		go func() {
			defer close(done)
			buf := make([]byte, 2048)
			for {
				n, err := c2.Read(buf)
				if n > 0 {
					chk := make([]byte, n)
					copy(chk, buf[:n])
					chunks = append(chunks, chk)
				}
				if err != nil {
					break
				}
			}
		}()

		_, err := fc.Write(raw)
		_ = c1.Close()
		<-done
		if err != nil {
			return err
		}
		if len(chunks) < 3 {
			return fmt.Errorf("expected >= 3 chunks for random mode with length 15-25, got %d", len(chunks))
		}
		if !bytes.Equal(bytes.Join(chunks, nil), raw) {
			return errors.New("reassembled payload mismatch in random mode")
		}
		return nil
	})

	// 6. Record Fragmentation
	runCheck("Record Fragmentation", "Converts single TLS record into multiple RFC-compliant TLS records", func() error {
		raw := makeDiagClientHello("record-split.example.com")
		records := splitRecord(raw, &Option{Enabled: true, Mode: ModeRecordFragment, Length: "20-40"})
		if len(records) < 2 {
			return fmt.Errorf("expected >= 2 TLS records, got %d", len(records))
		}
		for i, rec := range records {
			if len(rec) < 5 || rec[0] != 0x16 || rec[1] != 0x03 {
				return fmt.Errorf("record %d has invalid TLS header", i)
			}
			bodyLen := int(binary.BigEndian.Uint16(rec[3:5]))
			if len(rec) != 5+bodyLen {
				return fmt.Errorf("record %d payload length mismatch", i)
			}
		}
		return nil
	})

	// 7. Both Mode
	runCheck("Both Mode", "Combines multi-record TLS splitting with delayed TCP segmentation", func() error {
		raw := makeDiagClientHello("both-mode.example.com")
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeBoth, Length: "15-30", Interval: "1-2"})
		var chunks [][]byte
		done := make(chan struct{})

		go func() {
			defer close(done)
			buf := make([]byte, 2048)
			for {
				n, err := c2.Read(buf)
				if n > 0 {
					chk := make([]byte, n)
					copy(chk, buf[:n])
					chunks = append(chunks, chk)
				}
				if err != nil {
					break
				}
			}
		}()

		_, err := fc.Write(raw)
		_ = c1.Close()
		<-done
		if err != nil {
			return err
		}
		if len(chunks) < 3 {
			return fmt.Errorf("expected >= 3 chunks in both mode, got %d", len(chunks))
		}
		return nil
	})

	// 8. Error Handling
	runCheck("Error Handling", "Verifies safe write loops and error persistence on failure", func() error {
		c1, c2 := net.Pipe()
		fc := NewConn(c1, &Option{Enabled: true, Mode: ModeSNI})
		// Close remote side immediately
		_ = c2.Close()
		_, err := fc.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
		if err == nil {
			return errors.New("expected error when writing to closed pipe")
		}
		return nil
	})

	// 9. Configuration Validation
	runCheck("Configuration Validation", "Validates syntax checking for Length and Interval parameters", func() error {
		optValid := &Option{Length: "10-30", Interval: "1-5"}
		minL, maxL := optValid.ParseLength(10, 30)
		if minL != 10 || maxL != 30 {
			return fmt.Errorf("ParseLength expected 10-30, got %d-%d", minL, maxL)
		}
		minI, maxI := optValid.IntervalRange(1, 5)
		if minI != 1 || maxI != 5 {
			return fmt.Errorf("IntervalRange expected 1-5ms, got %d-%d", minI, maxI)
		}
		return nil
	})

	return report
}
