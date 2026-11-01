// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).

package fragment

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestStats_CountersAndReset(t *testing.T) {
	ResetGlobalStats()

	s := GetGlobalStats()
	if s.TotalClientHello != 0 || s.SuccessfulFragments != 0 || s.FailedFragments != 0 {
		t.Fatalf("expected zero stats after reset, got %+v", s)
	}

	RecordClientHelloDetected("sni")
	RecordClientHelloDetected("random")
	RecordFragmentationSuccess("sni")
	RecordFragmentationSuccess("random")
	RecordFragmentationError("sni", errors.New("test connection error"))
	RecordBufferTimeout()
	RecordNonTLSBypass()

	s2 := GetGlobalStats()
	if s2.TotalClientHello != 2 {
		t.Errorf("expected TotalClientHello=2, got %d", s2.TotalClientHello)
	}
	if s2.SuccessfulFragments != 2 {
		t.Errorf("expected SuccessfulFragments=2, got %d", s2.SuccessfulFragments)
	}
	if s2.FailedFragments != 1 {
		t.Errorf("expected FailedFragments=1, got %d", s2.FailedFragments)
	}
	if s2.LastError != "test connection error" {
		t.Errorf("expected LastError='test connection error', got %s", s2.LastError)
	}
	if s2.Timeouts != 1 {
		t.Errorf("expected Timeouts=1, got %d", s2.Timeouts)
	}
	if s2.BypassedNonTLS != 1 {
		t.Errorf("expected BypassedNonTLS=1, got %d", s2.BypassedNonTLS)
	}
	if s2.ModeDistribution["sni"] != 1 || s2.ModeDistribution["random"] != 1 {
		t.Errorf("unexpected ModeDistribution: %+v", s2.ModeDistribution)
	}

	data, err := json.Marshal(s2)
	if err != nil {
		t.Fatalf("failed to marshal stats to JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty json stats output")
	}

	ResetGlobalStats()
	s3 := GetGlobalStats()
	if s3.TotalClientHello != 0 || s3.SuccessfulFragments != 0 {
		t.Fatalf("stats not cleared after reset: %+v", s3)
	}
}

func TestDiagnostics_RunDiagnostics(t *testing.T) {
	report := RunDiagnostics(nil)
	if !report.AllPassed {
		t.Fatalf("expected all diagnostics to pass, but AllPassed is false")
	}
	if len(report.Checks) != 9 {
		t.Fatalf("expected 9 diagnostic checks, got %d", len(report.Checks))
	}
	for _, check := range report.Checks {
		if check.Status != "PASS" {
			t.Errorf("check %s failed: %s (status: %s)", check.Name, check.Details, check.Status)
		}
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal diagnostics report to JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty json diagnostics output")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := &rateLimiter{maxPerSec: 5}
	allowed := 0
	for i := 0; i < 20; i++ {
		if rl.allow() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("expected 5 allowed calls in same second, got %d", allowed)
	}
	if rl.suppressed != 15 {
		t.Fatalf("expected 15 suppressed calls, got %d", rl.suppressed)
	}

	time.Sleep(1050 * time.Millisecond)
	if !rl.allow() {
		t.Fatal("expected allow() to return true in new second")
	}
}
