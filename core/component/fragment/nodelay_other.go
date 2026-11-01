// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev / mihomo).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

//go:build !windows

package fragment

import (
	"net"
	"syscall"
)

func setTCPNoDelayFallback(c net.Conn) {
	if sc, ok := c.(syscall.Conn); ok {
		if rc, err := sc.SyscallConn(); err == nil {
			_ = rc.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
			})
		}
	}
}
