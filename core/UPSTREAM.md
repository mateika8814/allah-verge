# Mihomo Core Upstream Reference

This directory contains the Mihomo core source code modified with TLS ClientHello packet and record fragmentation support (anti-DPI).

## Upstream Details
- **Repository**: [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo)
- **Base Commit**: `88dcbf7f1614a67c3b36b848ee3592dfa92ada36`
- **Base Commit Date**: `2026-09-30 22:35:49 +0800`
- **Base Commit Subject**: `fix: default listener tun to mips stack (#3264)`
- **License**: GNU General Public License v3.0 (see [LICENSE](./LICENSE))

## Modifications in this Fork
Our anti-DPI fragmentation engine is implemented in `component/fragment/` and hooked into outbound dialers in `component/dialer/` and `adapter/outbound/`:
1. `component/fragment/conn.go` - `fragConn` net.Conn wrapper implementing packet and record fragmentation with buffering and interval delays.
2. `component/fragment/option.go` - Configuration structures and range parsers (`sni`, `random`, `record_fragment`, `both`). Includes unknown mode fallback warning logging.
3. `component/fragment/sni.go` - RFC-compliant TLS ClientHello parser identifying exact SNI offsets without external dependencies.
4. `component/fragment/conn_test.go` & `sni_test.go` - Unit test suites verifying segmentation boundaries and TLS record framing.
5. `component/dialer/dialer.go` & `options.go` - Fragment wrapper hook in TCP outbound dialing pipeline.
6. `adapter/outbound/base.go` - Parsing proxy-level `fragment:` configuration options from YAML configs.

All modifications relative to the upstream base commit are recorded in [`fragment.patch`](./fragment.patch).

## How to Build the Core
From this directory:
```bash
go build -v -trimpath -ldflags "-s -w -buildid=" -tags with_gvisor -o verge-mihomo.exe .
```
The resulting `verge-mihomo.exe` binary can be placed in Clash Verge's portable root or sidecar folder.
