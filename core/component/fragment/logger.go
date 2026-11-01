// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fragment

import (
	"fmt"
	"sync"
	"time"

	"github.com/metacubex/mihomo/log"
)

type rateLimiter struct {
	mu         sync.Mutex
	lastSec    int64
	count      int
	maxPerSec  int
	suppressed int
}

var infoLimiter = &rateLimiter{
	maxPerSec: 15, // up to 15 info events per second
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	nowSec := time.Now().Unix()
	if nowSec != rl.lastSec {
		if rl.suppressed > 0 {
			log.Infoln("[Fragment][INFO] (... %d more fragment events in last second)", rl.suppressed)
			rl.suppressed = 0
		}
		rl.lastSec = nowSec
		rl.count = 0
	}

	if rl.count < rl.maxPerSec {
		rl.count++
		return true
	}
	rl.suppressed++
	return false
}

// LogInfo logs an informational message if rate limiter allows.
func LogInfo(format string, args ...any) {
	if infoLimiter.allow() {
		log.Infoln("[Fragment][INFO] "+format, args...)
	}
}

// LogDebug logs a debug message (no rate limit, debug level only).
func LogDebug(format string, args ...any) {
	log.Debugln("[Fragment][DEBUG] "+format, args...)
}

// LogWarn logs a warning message.
func LogWarn(format string, args ...any) {
	log.Warnln("[Fragment][WARNING] "+format, args...)
}

// LogError logs an error message.
func LogError(format string, args ...any) {
	log.Errorln("[Fragment][ERROR] "+format, args...)
}

// TargetIdentifier builds a masked, privacy-friendly connection descriptor.
func TargetIdentifier(connRemote string, sni string) string {
	maskedSNI := MaskDomain(sni)
	if maskedSNI != "" && connRemote != "" {
		return fmt.Sprintf("%s (%s)", maskedSNI, connRemote)
	}
	if maskedSNI != "" {
		return maskedSNI
	}
	if connRemote != "" {
		return connRemote
	}
	return "unknown"
}
