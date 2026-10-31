import { describe, expect, it } from 'vitest'

import { isTlsProxy, validateIntervalRange, validateLengthRange } from './fragment'

describe('TLS Fragmentation validation & logic', () => {
  describe('validateLengthRange', () => {
    it('accepts valid single numbers >= 1', () => {
      expect(validateLengthRange('1')).toBe(true)
      expect(validateLengthRange('10')).toBe(true)
      expect(validateLengthRange('100')).toBe(true)
      expect(validateLengthRange('  25  ')).toBe(true)
    })

    it('accepts valid ranges where N <= M and N >= 1', () => {
      expect(validateLengthRange('1-1')).toBe(true)
      expect(validateLengthRange('10-30')).toBe(true)
      expect(validateLengthRange('100-200')).toBe(true)
      expect(validateLengthRange('  5 - 15  ')).toBe(true)
    })

    it('rejects 0 or negative lengths', () => {
      expect(validateLengthRange('0')).toBe(false)
      expect(validateLengthRange('0-10')).toBe(false)
      expect(validateLengthRange('-5')).toBe(false)
      expect(validateLengthRange('-10--5')).toBe(false)
    })

    it('rejects inverted ranges where N > M', () => {
      expect(validateLengthRange('30-10')).toBe(false)
      expect(validateLengthRange('5-2')).toBe(false)
    })

    it('rejects non-numeric and malformed strings', () => {
      expect(validateLengthRange('')).toBe(false)
      expect(validateLengthRange('abc')).toBe(false)
      expect(validateLengthRange('10-abc')).toBe(false)
      expect(validateLengthRange('10-20-30')).toBe(false)
      expect(validateLengthRange('10.5-20')).toBe(false)
    })
  })

  describe('validateIntervalRange', () => {
    it('accepts valid single numbers >= 0', () => {
      expect(validateIntervalRange('0')).toBe(true)
      expect(validateIntervalRange('5')).toBe(true)
      expect(validateIntervalRange('  10  ')).toBe(true)
    })

    it('accepts valid ranges where N <= M and N >= 0', () => {
      expect(validateIntervalRange('0-0')).toBe(true)
      expect(validateIntervalRange('0-5')).toBe(true)
      expect(validateIntervalRange('1-5')).toBe(true)
      expect(validateIntervalRange('  2 - 8  ')).toBe(true)
    })

    it('rejects negative intervals', () => {
      expect(validateIntervalRange('-1')).toBe(false)
      expect(validateIntervalRange('-5-5')).toBe(false)
    })

    it('rejects inverted ranges where N > M', () => {
      expect(validateIntervalRange('5-1')).toBe(false)
      expect(validateIntervalRange('10-0')).toBe(false)
    })

    it('rejects malformed input', () => {
      expect(validateIntervalRange('')).toBe(false)
      expect(validateIntervalRange('xyz')).toBe(false)
      expect(validateIntervalRange('1-2-3')).toBe(false)
    })
  })

  describe('isTlsProxy', () => {
    it('correctly identifies inherent TLS protocols', () => {
      expect(isTlsProxy({ type: 'trojan' })).toBe(true)
      expect(isTlsProxy({ type: 'anytls' })).toBe(true)
      expect(isTlsProxy({ type: 'tuic' })).toBe(true)
      expect(isTlsProxy({ type: 'hysteria' })).toBe(true)
      expect(isTlsProxy({ type: 'hysteria2' })).toBe(true)
    })

    it('correctly identifies vless with TLS or Reality', () => {
      expect(isTlsProxy({ type: 'vless' })).toBe(true)
      expect(isTlsProxy({ type: 'vless', security: 'tls' })).toBe(true)
      expect(isTlsProxy({ type: 'vless', security: 'reality' })).toBe(true)
      expect(isTlsProxy({ type: 'vless', 'reality-opts': { 'public-key': 'abc' } })).toBe(true)
      expect(isTlsProxy({ type: 'vless', security: 'none' })).toBe(false)
    })

    it('correctly identifies vmess with/without TLS', () => {
      expect(isTlsProxy({ type: 'vmess' })).toBe(false)
      expect(isTlsProxy({ type: 'vmess', tls: false })).toBe(false)
      expect(isTlsProxy({ type: 'vmess', tls: true })).toBe(true)
      expect(isTlsProxy({ type: 'vmess', security: 'tls' })).toBe(true)
    })

    it('correctly identifies non-TLS proxies', () => {
      expect(isTlsProxy({ type: 'direct' })).toBe(false)
      expect(isTlsProxy({ type: 'dns' })).toBe(false)
      expect(isTlsProxy({ type: 'ss' })).toBe(false)
      expect(isTlsProxy(null)).toBe(false)
      expect(isTlsProxy(undefined)).toBe(false)
    })

    it('identifies shadowsocks with TLS plugin', () => {
      expect(
        isTlsProxy({
          type: 'ss',
          plugin: 'v2ray-plugin',
          'plugin-opts': { mode: 'websocket', tls: true },
        }),
      ).toBe(true)
      expect(isTlsProxy({ type: 'ss', plugin: 'obfs-local' })).toBe(false)
    })
  })

  describe('Contract and Merge Logic', () => {
    it('omits length when mode is sni, retains length when random', () => {
      const buildFragment = (
        mode: 'sni' | 'random',
        length: string,
        interval: string,
      ) => {
        return {
          enabled: true,
          mode,
          ...(mode === 'random' ? { length } : {}),
          interval,
          'record-fragment': false,
        }
      }

      const sniFrag = buildFragment('sni', '10-30', '1-5')
      expect(sniFrag.mode).toBe('sni')
      expect('length' in sniFrag).toBe(false)
      expect(sniFrag.interval).toBe('1-5')

      const randFrag = buildFragment('random', '10-30', '1-5')
      expect(randFrag.mode).toBe('random')
      expect(randFrag.length).toBe('10-30')
      expect(randFrag.interval).toBe('1-5')
    })
  })
})
