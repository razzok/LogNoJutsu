# Phase 17: Network Discovery - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-10
**Phase:** 17-network-discovery
**Areas discussed:** Scan scope & targeting, Port & protocol selection, T1018 discovery methods, Sysmon artifact quality

---

## Scan Scope & Targeting

### Q1: How should T1046 determine its scan target range?

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-detect /24 only | Reuse detectLocalSubnet() from Phase 16. Safest default. | ✓ |
| YAML-configurable CIDR | Add target_range field to technique YAML with fallback. | |
| You decide | Claude picks approach. | |

**User's choice:** Auto-detect /24 only
**Notes:** Matches scan confirmation modal which already shows this subnet.

### Q2: Should the scanner exclude the host's own IP?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, exclude own IP | Skip host's own address — self-scan not interesting for SIEM. | ✓ |
| No, scan everything | Include own IP for completeness. | |
| You decide | Claude picks based on SIEM artifact quality. | |

**User's choice:** Yes, exclude own IP

### Q3: Rate limiting strategy?

| Option | Description | Selected |
|--------|-------------|----------|
| 50 conn/sec (match Phase 16 UI) | Phase 16 modal already shows this value. | ✓ |
| Configurable via input_args | Default 50/sec with YAML override. | |
| No rate limit (full speed) | Real scanners go fast. | |

**User's choice:** 50 conn/sec (match Phase 16 UI)

---

## Port & Protocol Selection

### Q4: Default port set for T1046?

| Option | Description | Selected |
|--------|-------------|----------|
| Common services (~20 ports) | 21,22,23,25,53,80,135,139,389,443,445,1433,3306,3389,5985,8080,8443. | ✓ |
| Top 100 ports | Nmap-style top 100 TCP ports. | |
| Minimal (5 ports) | Just 80,443,445,3389,22. | |

**User's choice:** Common services (~20 ports)

### Q5: TCP only or TCP + UDP?

| Option | Description | Selected |
|--------|-------------|----------|
| TCP only | Clean Sysmon EID 3 events. | |
| TCP + limited UDP | Add UDP for DNS(53), SNMP(161), NTP(123). | ✓ |

**User's choice:** TCP + limited UDP

---

## T1018 Discovery Methods

### Q6: Which discovery methods for T1018?

| Option | Description | Selected |
|--------|-------------|----------|
| All four | ICMP ping sweep, ARP table dump, nltest DC discovery, DNS reverse lookups. | ✓ |
| Ping + ARP only | Simpler, skip nltest/DNS. | |
| Ping + ARP + nltest | Skip DNS reverse lookups. | |

**User's choice:** All four

### Q7: Single technique or split sub-techniques?

| Option | Description | Selected |
|--------|-------------|----------|
| Single technique, multiple methods | One T1018 YAML with type: go running all methods sequentially. | ✓ |
| Split into sub-techniques | T1018.001, T1018.002, T1018.003. | |

**User's choice:** Single technique, multiple methods

---

## Sysmon Artifact Quality

### Q8: Goroutine concurrency for EID 3 burst?

| Option | Description | Selected |
|--------|-------------|----------|
| Goroutine pool (10-20 concurrent) | Worker pool with net.DialTimeout(). Burst signature from one PID. | ✓ |
| Sequential with small delay | One connection at a time with 20ms delay. | |
| You decide | Claude picks concurrency model. | |

**User's choice:** Goroutine pool (10-20 concurrent)

### Q9: ICMP ping sweep approach?

| Option | Description | Selected |
|--------|-------------|----------|
| net.Dial('ip4:icmp') with fallback | Try ICMP first (needs admin), fall back to TCP on 445/135 if not admin. | ✓ |
| TCP-only host alive check | Skip ICMP entirely, use TCP SYN on common ports. | |
| You decide | Claude picks based on artifact quality. | |

**User's choice:** net.Dial('ip4:icmp') with fallback

---

## Claude's Discretion

- Connection timeout values, UDP implementation details, ARP table reading approach, nltest wrapper, DNS reverse lookup implementation, worker pool size within 10-20, output formatting, test strategy

## Deferred Ideas

None — discussion stayed within phase scope.
