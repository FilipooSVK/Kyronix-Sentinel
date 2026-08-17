# Kyronix Sentinel v0.1.1

<p align="center">
  <strong>Predictive Linux host health monitoring, degradation detection and safe recovery intelligence.</strong>
</p>

<p align="center">
  Built by <strong>Kyronix</strong> for Linux infrastructure, Raspberry Pi, homelabs, appliances and resilient operations.
</p>

<p align="center">
  <a href="https://ko-fi.com/X8X31QYP4A">
    <img src="https://ko-fi.com/img/githubbutton_sm.svg" alt="ko-fi" />
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Architecture-ARM64%20%7C%20AMD64-lightgrey?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Health-Predictive-orange?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Updates-GitHub%20Releases-6f42c1?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Status-Development-yellow?style=for-the-badge" />
</p>

---

## 🛡️ What is Kyronix Sentinel?

**Kyronix Sentinel** is a lightweight Linux host health agent designed to detect **system degradation before it becomes a failure**.

Instead of acting as another general-purpose monitoring platform, Sentinel focuses on a different question:

> Is this system gradually becoming unstable — and can we detect it before it freezes?

Sentinel continuously collects host health information, evaluates multiple independent signals, keeps historical context and produces an explainable:

- **Health Score**
- **Freeze Risk**
- **Prediction Confidence**
- **Recovery Recommendation**

The long-term goal is simple:

> Detect degradation early enough to take safe action before the host becomes unavailable.

---

## 🚀 Why it exists

Traditional monitoring is excellent at showing what a system looks like **right now**.

But many failures do not happen instantly.

Before a Linux host freezes or becomes unstable, there may already be warning signs:

- available memory slowly decreasing,
- memory pressure increasing,
- I/O pressure growing,
- swap activity becoming abnormal,
- kernel OOM events appearing,
- filesystem problems accumulating,
- health scores degrading across multiple cycles,
- multiple weak signals appearing at the same time.

Individually, these signals may not justify intervention.

Together — and especially when persistent — they can indicate that the system is moving toward an unhealthy state.

**Kyronix Sentinel is built to recognize that pattern.**

---

## ✨ Features

- 🧠 Predictive system health analysis
- ❤️ Host Health Score
- ❄️ Freeze Risk estimation
- 📈 Historical trend analysis
- 🧩 Multi-signal consensus engine
- 💾 Persistent health history
- 🧠 Memory degradation detection
- ⚡ Linux PSI pressure monitoring
- 💽 Disk and filesystem health signals
- ☠️ Kernel and cgroup OOM detection
- 🔍 Explainable prediction reasons
- 🧰 Local `sentinelctl` management CLI
- 🔌 Local Unix socket API
- 🔄 GitHub Release update detection
- 🔐 SHA256 update verification
- 📦 Release manifest validation
- ⚙️ Atomic binary installation
- 🛟 Automatic rollback infrastructure
- 🐧 Native systemd integration

---

## 🧱 Architecture

### Predictive health pipeline

```text
                   ┌─────────────────┐
                   │    Linux Host   │
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │    Collectors   │
                   └────────┬────────┘
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
          ▼                 ▼                 ▼
        CPU              Memory             Disk
          │                 │                 │
          ├────────── Pressure / PSI ─────────┤
          │                                   │
          └────────── Kernel / OOM ───────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │    Analyzer     │
                   └────────┬────────┘
                            │
                    Health Score
                     Freeze Risk
                            │
                            ▼
                   ┌─────────────────┐
                   │     History     │
                   │ Rolling + Disk  │
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │   Prediction    │
                   │     Engine      │
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │ Recommendation  │
                   └────────┬────────┘
                            │
                 ┌──────────┴──────────┐
                 │                     │
                 ▼                     ▼
           Unix Socket API        sentinelctl
```

---

## 🧠 Prediction model

Sentinel does not rely on a single metric.

The prediction engine evaluates multiple independent signals and considers whether they persist across time.

Current risk levels:

```text
LOW
MEDIUM
HIGH
CRITICAL
```

Current recommendations:

```text
MONITOR
INVESTIGATE
REBOOT_ADVISED
AUTO_RECOVERY
```

> `AUTO_RECOVERY` currently represents a recommendation only. Automatic host reboot is not enabled.

This is intentional.

Before Sentinel is allowed to perform disruptive recovery actions, the recovery safety model must be fully validated.

---

## 📊 Example prediction

```text
Kyronix Sentinel Prediction

Risk: LOW
Score: 10
Confidence: 90%
Recommendation: MONITOR

Consensus:
  Active signals: 0
  Persistent signals: 0
  Kernel evidence: no

Reasons:
  - slow memory growth
```

Sentinel is designed to remain **explainable**.

A risk score should never appear without information about why the decision was made.

---

## ❤️ Health Score

The Health Score provides a simplified representation of the current system condition.

Example:

```text
Kyronix Sentinel

Running: true
Health: 90
Risk: LOW
```

The score is influenced by health findings produced by Sentinel's analyzer.

It is not intended to replace detailed monitoring metrics.

Instead, it provides a fast operational answer to:

> How healthy does this host currently appear?

---

## 🔎 Current health signals

Sentinel currently evaluates information from several Linux subsystems.

### Memory

- Available memory
- Swap state
- Memory utilization
- Memory growth trends

### Pressure

Linux **Pressure Stall Information (PSI)** where supported:

```text
/proc/pressure/cpu
/proc/pressure/memory
/proc/pressure/io
```

PSI is optional.

If the host does not expose pressure information, Sentinel continues operating without PSI-based signals.

### Kernel

- System OOM kills
- cgroup OOM events
- cgroup OOM kills
- filesystem error indicators

### Disk

- Mounted filesystems
- Storage utilization
- Disk-related health findings

### CPU

- CPU utilization
- CPU collection state
- Historical CPU behavior

---

## 💾 Persistent history

Sentinel stores historical snapshots so that prediction does not depend only on the current collection cycle.

Default history location:

```text
/var/lib/sentinel/history.jsonl
```

Historical data allows Sentinel to distinguish between:

```text
temporary spike
```

and:

```text
persistent degradation
```

The history survives daemon restarts and uses automatic retention and compaction.

---

## ⚙️ Components

### `sentineld`

The background health and prediction daemon.

Responsibilities include:

- collecting system metrics,
- evaluating health findings,
- maintaining historical state,
- calculating health score,
- estimating freeze risk,
- generating predictions,
- exposing the local management API.

---

### `sentinelctl`

The local Sentinel management CLI.

Current commands include:

```bash
sentinelctl version
sentinelctl status
sentinelctl diagnose
sentinelctl prediction
sentinelctl update check
sudo sentinelctl update install
```

---

## 🔌 Local API

Sentinel exposes a local Unix socket:

```text
/run/sentinel/sentinel.sock
```

The API provides runtime information including:

- daemon state,
- Sentinel version,
- Health Score,
- Freeze Risk,
- prediction data,
- collector diagnostics.

The API is also used by the update system to verify that a newly installed Sentinel version started successfully.

---

## 🔧 Configuration

Default configuration:

```text
/etc/sentinel/sentinel.yaml
```

Example:

```yaml
daemon:
  interval: 30s

history:
  size: 1000
  persistence: true
  path: /var/lib/sentinel/history.jsonl

logging:
  level: info

update:
  enabled: true
  owner: FilipooSVK
  repository: Kyronix-Sentinel
  auto_check: true
  auto_install: false
  check_interval: 24h
```

Automatic update installation remains disabled by default.

---

## 🔄 Secure update system

Kyronix Sentinel includes its own GitHub Release update infrastructure.

Check for a newer release:

```bash
sentinelctl update check
```

Example:

```text
Kyronix Sentinel Update

Current: v0.1.0
Latest: v0.1.1
Status: UPDATE AVAILABLE
```

Install an available update:

```bash
sudo sentinelctl update install
```

---

### Update pipeline

```text
GitHub Releases
      │
      ▼
Check latest version
      │
      ▼
Select Linux architecture
      │
      ├── ARM64
      └── AMD64
      │
      ▼
Download checksum
      │
      ▼
Download package
      │
      ▼
Verify SHA256
      │
      ▼
Safe archive extraction
      │
      ▼
Validate manifest
      │
      ▼
Stage new binaries
      │
      ▼
Backup current binaries
      │
      ▼
Atomic installation
      │
      ▼
Restart sentineld
      │
      ▼
Local API health check
      │
      ▼
Verify expected version
      │
      ▼
UPDATE SUCCESS
```

---

## 🔐 Update security

Sentinel release archives contain only:

```text
manifest.json
sentineld
sentinelctl
```

The manifest defines:

- release version,
- target operating system,
- target architecture,
- expected binary files,
- SHA256 hash for each binary.

Example assets:

```text
sentinel-v0.1.1-linux-arm64.tar.gz
sentinel-v0.1.1-linux-arm64.tar.gz.sha256

sentinel-v0.1.1-linux-amd64.tar.gz
sentinel-v0.1.1-linux-amd64.tar.gz.sha256
```

Unsafe archive contents are rejected, including:

- path traversal,
- absolute paths,
- symbolic links,
- hard links,
- unsupported filesystem objects.

---

## 🛟 Rollback protection

Before replacing the running Sentinel binaries, the updater keeps the previous versions as:

```text
/usr/local/bin/sentineld.previous
/usr/local/bin/sentinelctl.previous
```

If activation or the post-install health check fails, the updater contains logic to restore the previous binaries and restart Sentinel.

The basic rollback infrastructure is implemented.

> Real-world controlled failure and rollback validation is still part of the development roadmap.

---

## ✅ First successful self-update

The first complete end-to-end Sentinel update was successfully validated with:

```text
v0.1.0
   │
   ▼
GitHub Release v0.1.1
   │
   ▼
ARM64 package selected
   │
   ▼
SHA256 verified
   │
   ▼
Manifest validated
   │
   ▼
Atomic install
   │
   ▼
sentineld restarted
   │
   ▼
Local health check passed
   │
   ▼
Version verified
   │
   ▼
v0.1.1
```

After the update:

```text
Current: v0.1.1
Latest: v0.1.1
Status: UP TO DATE
```

---

## 🐧 Supported platforms

| Platform | Architecture | Status |
| -------- | ------------ | ------ |
| Linux    | ARM64        | ✅ Supported |
| Linux    | AMD64        | ✅ Supported |

Primary development currently focuses on:

- Debian
- Ubuntu
- Raspberry Pi OS / Debian-based ARM64 systems

Future environments include:

- Proxmox VE
- Docker hosts
- infrastructure appliances
- Kyronix Stratus OS

---

## 🛠️ Building from source

Requirements:

- Linux
- Go 1.24.x

Clone:

```bash
git clone git@github.com:FilipooSVK/Kyronix-Sentinel.git
cd Kyronix-Sentinel
```

Run tests:

```bash
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go vet ./...
```

Build Sentinel:

```bash
mkdir -p bin

GOTOOLCHAIN=local go build \
  -o bin/sentineld \
  ./cmd/sentineld

GOTOOLCHAIN=local go build \
  -o bin/sentinelctl \
  ./cmd/sentinelctl
```

---

## 📦 Building a release

Sentinel includes its own release builder.

Example:

```bash
./scripts/build-release.sh 0.1.1
```

Generated artifacts are stored under:

```text
dist/v0.1.1/
```

The release builder automatically:

- builds ARM64 binaries,
- builds AMD64 binaries,
- embeds version metadata,
- embeds Git commit information,
- generates release manifests,
- calculates binary SHA256 hashes,
- creates compressed release archives,
- generates external archive checksums.

---

## ⚙️ systemd

Sentinel runs as a native systemd service.

Service:

```text
sentineld.service
```

Check its state:

```bash
systemctl status sentineld
```

Quick status:

```bash
systemctl is-active sentineld
```

---

## 🧭 Design principles

Kyronix Sentinel follows several core principles.

### 🧠 Explainable decisions

Every significant prediction should have observable reasons.

---

### 🧩 Multi-signal consensus

One abnormal metric should not normally trigger a disruptive recovery action.

---

### 📈 History matters

Persistent degradation is more important than a temporary spike.

---

### 🛡️ Conservative recovery

Rebooting a production host is disruptive.

Sentinel must be confident before recommending — and eventually performing — recovery.

---

### 🔐 Safe updates

Software updates must be verified before installation and recoverable when activation fails.

---

### 🪶 Lightweight by design

Sentinel is not intended to replace:

- Prometheus
- Grafana
- Zabbix
- enterprise observability platforms

Instead, it provides lightweight **local host intelligence** focused specifically on degradation and recovery prediction.

---

## 🗺️ Roadmap

### Completed

- [x] Linux host collectors
- [x] CPU collector
- [x] Memory collector
- [x] Disk collector
- [x] Kernel / OOM collector
- [x] Linux PSI support
- [x] Health Score
- [x] Freeze Risk
- [x] Historical trend analysis
- [x] Persistent history
- [x] Multi-signal prediction consensus
- [x] Prediction explainability
- [x] Local Unix socket API
- [x] `sentinelctl`
- [x] GitHub Release discovery
- [x] Architecture-aware update selection
- [x] SHA256 release verification
- [x] Safe archive extraction
- [x] Manifest validation
- [x] Atomic binary installation
- [x] Post-install health verification
- [x] Rollback infrastructure
- [x] First successful end-to-end self-update

### Next

- [ ] Controlled real-world rollback test
- [ ] Periodic automatic update checks
- [ ] Persistent update state
- [ ] Controlled automatic update policy
- [ ] Recovery safety gates
- [ ] Automatic recovery execution
- [ ] Predictive reboot policy
- [ ] Proxmox-specific intelligence
- [ ] Docker host intelligence

---

## ⚠️ Development status

Kyronix Sentinel is currently an **experimental infrastructure project**.

Version `0.x` releases should be considered development releases.

Automatic reboot and unattended update installation are intentionally not enabled by default.

Do not rely on Sentinel as the sole mechanism protecting critical production infrastructure.

---

## 🤝 Contributing

Suggestions, testing and technical discussions are welcome through **GitHub Issues**.

Useful contribution areas include:

- Linux hardware testing
- ARM64 testing
- AMD64 testing
- failure scenarios
- kernel health signals
- prediction heuristics
- Proxmox testing
- container-host health detection
- documentation improvements

---

## 📦 Releases

Official Kyronix Sentinel builds are published through GitHub Releases:

**https://github.com/FilipooSVK/Kyronix-Sentinel/releases**

Current release:

```text
v0.1.1
```

---

## 🧭 Project

Built by **Kyronix**.

Kyronix Sentinel is part of a broader infrastructure engineering project focused on:

- Predictive operations
- Infrastructure health
- Observability
- Automated recovery
- Secure software lifecycle
- Resilient Linux systems

> **Observe early. Predict degradation. Recover safely.**
