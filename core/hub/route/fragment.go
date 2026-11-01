// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).

package route

import (
	"encoding/json"

	"github.com/metacubex/chi"
	"github.com/metacubex/chi/render"
	"github.com/metacubex/http"
	"github.com/metacubex/mihomo/component/fragment"
)

func fragmentRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/stats", getFragmentStats)
	r.Post("/stats/reset", resetFragmentStats)
	r.Get("/diagnostics", getFragmentDiagnostics)
	r.Post("/diagnostics", runFragmentDiagnostics)
	return r
}

func getFragmentStats(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, fragment.GetGlobalStats())
}

func resetFragmentStats(w http.ResponseWriter, r *http.Request) {
	fragment.ResetGlobalStats()
	render.JSON(w, r, render.M{"status": "ok"})
}

func getFragmentDiagnostics(w http.ResponseWriter, r *http.Request) {
	report := fragment.RunDiagnostics(nil)
	render.JSON(w, r, report)
}

func runFragmentDiagnostics(w http.ResponseWriter, r *http.Request) {
	var opt fragment.Option
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&opt)
	}
	report := fragment.RunDiagnostics(&opt)
	render.JSON(w, r, report)
}
