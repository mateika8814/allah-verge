// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fragment

import (
	"sync"
	"time"
)

// Stats holds real-time counters and diagnostics metrics for TLS fragmentation.
type Stats struct {
	TotalClientHello    int64            `json:"total_client_hello"`
	SuccessfulFragments int64            `json:"successful_fragments"`
	FailedFragments     int64            `json:"failed_fragments"`
	Timeouts            int64            `json:"timeouts"`
	BypassedNonTLS      int64            `json:"bypassed_non_tls"`
	ModeDistribution    map[string]int64 `json:"mode_distribution"`
	LastError           string           `json:"last_error"`
	LastErrorTime       int64            `json:"last_error_time"`
	LastEventTime       int64            `json:"last_event_time"`
}

type statsManager struct {
	mu    sync.RWMutex
	stats Stats
}

var globalStats = &statsManager{
	stats: Stats{
		ModeDistribution: make(map[string]int64),
	},
}

// GetGlobalStats returns a copy of current fragment statistics.
func GetGlobalStats() Stats {
	globalStats.mu.RLock()
	defer globalStats.mu.RUnlock()

	res := globalStats.stats
	res.ModeDistribution = make(map[string]int64, len(globalStats.stats.ModeDistribution))
	for k, v := range globalStats.stats.ModeDistribution {
		res.ModeDistribution[k] = v
	}
	return res
}

// ResetGlobalStats clears all counters.
func ResetGlobalStats() {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats = Stats{
		ModeDistribution: make(map[string]int64),
	}
}

// RecordClientHelloDetected records a detected TLS ClientHello event.
func RecordClientHelloDetected(mode string) {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats.TotalClientHello++
	globalStats.stats.LastEventTime = time.Now().Unix()
}

// RecordFragmentationSuccess records an executed fragmentation.
func RecordFragmentationSuccess(mode string) {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats.SuccessfulFragments++
	globalStats.stats.ModeDistribution[mode]++
	globalStats.stats.LastEventTime = time.Now().Unix()
}

// RecordFragmentationError records a fragmentation failure.
func RecordFragmentationError(mode string, err error) {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats.FailedFragments++
	if err != nil {
		globalStats.stats.LastError = err.Error()
		globalStats.stats.LastErrorTime = time.Now().Unix()
	}
	globalStats.stats.LastEventTime = time.Now().Unix()
}

// RecordBufferTimeout records a buffering timeout event.
func RecordBufferTimeout() {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats.Timeouts++
	globalStats.stats.LastEventTime = time.Now().Unix()
}

// RecordNonTLSBypass records non-TLS traffic forwarded without fragmentation.
func RecordNonTLSBypass() {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.stats.BypassedNonTLS++
	globalStats.stats.LastEventTime = time.Now().Unix()
}
