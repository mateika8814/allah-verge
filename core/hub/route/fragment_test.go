// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).

package route

import (
	"encoding/json"
	"testing"

	"github.com/metacubex/http"
	"github.com/metacubex/http/httptest"
	"github.com/metacubex/mihomo/component/fragment"
)

func TestFragmentEndpoints(t *testing.T) {
	h := fragmentRouter()

	// 1. GET /stats
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var stats fragment.Stats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to decode stats json: %v", err)
	}

	// 2. POST /stats/reset
	reqReset := httptest.NewRequest(http.MethodPost, "/stats/reset", nil)
	wReset := httptest.NewRecorder()
	h.ServeHTTP(wReset, reqReset)
	if wReset.Code != 200 {
		t.Fatalf("expected status 200 for reset, got %d", wReset.Code)
	}

	// 3. GET /diagnostics
	reqDiag := httptest.NewRequest(http.MethodGet, "/diagnostics", nil)
	wDiag := httptest.NewRecorder()
	h.ServeHTTP(wDiag, reqDiag)
	if wDiag.Code != 200 {
		t.Fatalf("expected status 200 for diagnostics, got %d", wDiag.Code)
	}

	var report fragment.DiagnosticsReport
	if err := json.Unmarshal(wDiag.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode diagnostics json: %v", err)
	}
	if !report.AllPassed {
		t.Fatalf("expected AllPassed to be true in diagnostics")
	}
	if len(report.Checks) != 9 {
		t.Fatalf("expected 9 checks, got %d", len(report.Checks))
	}
}
