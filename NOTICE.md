# NOTICE

## Project Attribution

This software project, **allah-verge**, is an open-source fork based on and incorporating code from:

1. **Clash Verge Rev**
   - **Upstream Repository**: [https://github.com/clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev)
   - **Base Commit**: `ef784e459d31fe9d1946568b6fe47ac3a7e692b7` (2026-10-30, `docs: document complete development setup (#8457)`)
   - **Original Authors & Contributors**: The Clash Verge Rev Project and Contributors
   - **License**: GNU General Public License v3.0 (GPLv3) (see root [LICENSE](./LICENSE))

2. **Mihomo (MetaCubeX)**
   - **Upstream Repository**: [https://github.com/MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo)
   - **Base Commit**: `88dcbf7f1614a67c3b36b848ee3592dfa92ada36` (2026-09-30, `fix: default listener tun to mips stack (#3264)`)
   - **Original Authors & Contributors**: MetaCubeX and Contributors
   - **License**: GNU General Public License v3.0 (GPLv3) (see [core/LICENSE](./core/LICENSE))

---

## Fork Information & Modifications

**Forked by mateika8814 (https://github.com/mateika8814)**

### Key Modifications in this Fork:
- **TLS ClientHello Packet & Record Fragmentation (Anti-DPI / TSPU Bypass)**:
  - **Core (`core/`)**: Added `component/fragment` package implementing customizable ClientHello fragmentation (`sni`, `random`, `record_fragment`, `both`) with configurable fragment lengths, per-segment timing intervals, non-TLS 6-byte validation, buffering timeouts, and TLS record fragmentation; integrated into outbound connection dialers; added fallback warnings on unknown modes. Cross-platform build verified on Linux, macOS (Darwin), and Windows.
  - **GUI & Backend**: Added per-node and global TLS fragmentation settings dialogs (`ProxyFragmentDialog`, `FragmentCard`), runtime injection logic into generated core configurations, validation utilities, unit tests, and GUI about/attribution indicators.

---

## Disclaimer & Warranty

- This project is an independent fork and is **not affiliated with, maintained by, or endorsed by** the original authors or maintainers of Clash Verge Rev or MetaCubeX/mihomo.
- **NO WARRANTY**: This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License v3.0 for more details.
