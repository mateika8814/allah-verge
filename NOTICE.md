# NOTICE

## Project Attribution

This software project, **allah-verge**, is an open-source fork based on and incorporating code from:

1. **Clash Verge Rev**
   - **Upstream Repository**: [https://github.com/clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev)
   - **Original Authors & Contributors**: The Clash Verge Rev Project and Contributors
   - **License**: GNU General Public License v3.0 (GPLv3) (see root [LICENSE](./LICENSE))

2. **Mihomo (MetaCubeX)**
   - **Upstream Repository**: [https://github.com/MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo)
   - **Original Authors & Contributors**: MetaCubeX and Contributors
   - **License**: GNU General Public License v3.0 (GPLv3) (see [core/LICENSE](./core/LICENSE))

---

## Fork Information & Modifications

**Forked by mateika8814 (https://github.com/mateika8814)**

### Key Modifications in this Fork:
- **TLS ClientHello Packet & Record Fragmentation (Anti-DPI / TSPU Bypass)**:
  - **Core (`core/`)**: Added `component/fragment` package implementing customizable ClientHello fragmentation (`sni`, `random`, `record_fragment`, `both`) with configurable fragment lengths, timing intervals, and TLS record fragmentation; integrated into outbound connection dialers; added fallback warnings on unknown modes.
  - **GUI & Backend**: Added per-node and global TLS fragmentation settings dialogs (`ProxyFragmentDialog`, `FragmentCard`), runtime injection logic into generated core configurations, validation utilities, unit tests, and GUI about/attribution indicators.

---

## Disclaimer & Warranty

- This project is an independent fork and is **not affiliated with, maintained by, or endorsed by** the original authors or maintainers of Clash Verge Rev or MetaCubeX/mihomo.
- **NO WARRANTY**: This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License v3.0 for more details.
