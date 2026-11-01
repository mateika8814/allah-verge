// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

/**
 * Validate length range string: "N" or "N-M"
 * Requirements: integers, N <= M, min >= 1
 */
export function validateLengthRange(val: string): boolean {
  if (!val || typeof val !== 'string') return false
  const s = val.trim()
  if (!s) return false

  const parts = s.split('-').map((p) => p.trim())
  if (parts.length === 1 && parts[0] !== '') {
    const n = Number(parts[0])
    return Number.isInteger(n) && n >= 1
  }
  if (parts.length === 2 && parts[0] !== '' && parts[1] !== '') {
    const n1 = Number(parts[0])
    const n2 = Number(parts[1])
    return (
      Number.isInteger(n1) &&
      Number.isInteger(n2) &&
      n1 >= 1 &&
      n1 <= n2
    )
  }
  return false
}

/**
 * Validate interval range string: "N" or "N-M"
 * Requirements: integers, N <= M, min >= 0
 */
export function validateIntervalRange(val: string): boolean {
  if (!val || typeof val !== 'string') return false
  const s = val.trim()
  if (!s) return false

  const parts = s.split('-').map((p) => p.trim())
  if (parts.length === 1 && parts[0] !== '') {
    const n = Number(parts[0])
    return Number.isInteger(n) && n >= 0
  }
  if (parts.length === 2 && parts[0] !== '' && parts[1] !== '') {
    const n1 = Number(parts[0])
    const n2 = Number(parts[1])
    return (
      Number.isInteger(n1) &&
      Number.isInteger(n2) &&
      n1 >= 0 &&
      n1 <= n2
    )
  }
  return false
}

/**
 * Determine whether a proxy node utilizes TLS
 */
export function isTlsProxy(proxy: any): boolean {
  if (!proxy || typeof proxy !== 'object') return false

  const type = String(proxy.type || '').toLowerCase()
  if (['trojan', 'anytls', 'tuic', 'hysteria', 'hysteria2'].includes(type)) {
    return true
  }

  const p = proxy as Record<string, any>
  const security = String(p.security || '').toLowerCase()

  if (type === 'vless') {
    if (security === 'none') {
      return false
    }
    return (
      p.tls === true ||
      security === 'tls' ||
      security === 'reality' ||
      Boolean(p['reality-opts']) ||
      security === ''
    )
  }

  if (type === 'vmess' || type === 'http' || type === 'socks5') {
    return p.tls === true || security === 'tls' || security === 'reality'
  }

  if (type === 'ss' || type === 'shadowsocks') {
    const plugin = String(p.plugin || '').toLowerCase()
    if (plugin.includes('tls')) return true

    const pluginOpts = p['plugin-opts'] as Record<string, any> | undefined
    if (pluginOpts) {
      if (pluginOpts.tls === true) return true
      const mode = String(pluginOpts.mode || '').toLowerCase()
      if (mode.includes('tls')) return true
    }
    return false
  }

  return (
    p.tls === true ||
    security === 'tls' ||
    security === 'reality' ||
    Boolean(p['reality-opts'])
  )
}
