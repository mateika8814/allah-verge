// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fragment

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/metacubex/mihomo/log"
)

type Mode string

const (
	ModeSNI            Mode = "sni"
	ModeRandom         Mode = "random"
	ModeRecordFragment Mode = "record_fragment"
	ModeBoth           Mode = "both"
)

// Option defines the fragment configuration per proxy node
type Option struct {
	Enabled        bool          `proxy:"enabled,omitempty" yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Mode           Mode          `proxy:"mode,omitempty" yaml:"mode,omitempty" json:"mode,omitempty"`
	Length         string        `proxy:"length,omitempty" yaml:"length,omitempty" json:"length,omitempty"`
	Interval       string        `proxy:"interval,omitempty" yaml:"interval,omitempty" json:"interval,omitempty"`
	RecordFragment bool          `proxy:"record-fragment,omitempty" yaml:"record-fragment,omitempty" json:"record-fragment,omitempty"`
	BufferTimeout  time.Duration `proxy:"-" yaml:"-" json:"-"`
}

func (o *Option) GetMode() Mode {
	if o == nil || o.Mode == "" {
		return ModeSNI
	}
	m := strings.ToLower(strings.TrimSpace(string(o.Mode)))
	switch m {
	case "sni":
		return ModeSNI
	case "random":
		return ModeRandom
	case "record_fragment", "record-fragment":
		return ModeRecordFragment
	case "both":
		return ModeBoth
	default:
		log.Warnln("[Fragment] unknown fragmentation mode '%s', falling back to 'sni'", o.Mode)
		return ModeSNI
	}
}

// IsRecordFragment returns whether TLS record splitting should be performed.
// Conflict resolution:
// If record-fragment: true is specified with mode: sni or mode: random,
// record-fragment takes precedence and enables record splitting.
func (o *Option) IsRecordFragment() bool {
	if o == nil {
		return false
	}
	m := o.GetMode()
	return o.RecordFragment || m == ModeRecordFragment || m == ModeBoth
}

func (o *Option) ParseLength(defMin, defMax int) (int, int) {
	if o == nil || o.Length == "" {
		return defMin, defMax
	}
	return parseRange(o.Length, defMin, defMax, false)
}

func (o *Option) IntervalRange(defMin, defMax int) (int, int) {
	if o == nil || o.Interval == "" {
		return defMin, defMax
	}
	return parseRange(o.Interval, defMin, defMax, true)
}

func (o *Option) RandomInterval(min, max int) time.Duration {
	if min <= 0 && max <= 0 {
		return 0
	}
	ms := randomBetween(min, max)
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func (o *Option) ParseInterval(defMin, defMax int) time.Duration {
	min, max := o.IntervalRange(defMin, defMax)
	return o.RandomInterval(min, max)
}

func (o *Option) GetBufferTimeout() time.Duration {
	if o != nil && o.BufferTimeout > 0 {
		return o.BufferTimeout
	}
	return 2 * time.Second
}

func parseRange(s string, defMin, defMax int, allowZero bool) (int, int) {
	if s == "" {
		return defMin, defMax
	}
	minBound := 1
	if allowZero {
		minBound = 0
	}
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		v, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || v < minBound {
			return defMin, defMax
		}
		return v, v
	}
	v1, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	v2, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || v1 < minBound || v2 < minBound {
		return defMin, defMax
	}
	if v1 > v2 {
		v1, v2 = v2, v1
	}
	return v1, v2
}

func randomBetween(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}
