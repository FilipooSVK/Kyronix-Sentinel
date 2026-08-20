<h1 align="center">Kyronix Sentinel</h1>

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
  <img src="https://img.shields.io/badge/Status-Stable-yellow?style=for-the-badge" />
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

The goal is simple:

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
- 🔐 SHA256 package and manifest verification
- 📦 Architecture-aware release validation
- ⚙️ Atomic binary installation
- 🛟 Automatic rollback with post-rollback verification
- 🚫 Failed-release quarantine
- 🔒 Global update operation lock
- 🧾 Persistent update lifecycle state
- 🧠 Automatic install policy engine
- 👀 Observe-only automatic update mode
- ⚙️ Dedicated systemd update worker
- 🔁 Transactional worker-unit installation and rollback
- 🧩 Clean-host worker bootstrap and migration support
- 🐧 Native systemd integration
- 🏗️ Linux ARM64 and AMD64 release builds

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
Health: 100
Risk: LOW
```

The score is influenced by health findings produced by Sentinel's analyzer.

It is not intended to replace detailed monitoring metrics.

Instead, it provides a fast operational answer to:

> How healthy does this host currently appear?

---

## 🔎 Current health signals

Sentinel evaluates information from several Linux subsystems.

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
- exposing the local management API,
- performing periodic update checks,
- evaluating automatic-install policy,
- provisioning the update worker on clean or migrated hosts.

### `sentinelctl`

The local Sentinel management CLI.

Current commands include:

```text
sentinelctl version
sentinelctl status
sentinelctl diagnose
sentinelctl prediction

sentinelctl update check
sentinelctl update status
sentinelctl update policy
sentinelctl update quarantine

sudo sentinelctl update install
sudo sentinelctl update quarantine clear
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

Default production configuration:

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
  auto_install_mode: observe_only
  check_interval: 24h
  state_path: /var/lib/sentinel/update-state.json

  auto_install_policy:
    min_release_age: 24h
    patch_only: true
```

Automatic update installation remains **disabled by default**.

The recommended production mode for `v1.0.0` is:

```yaml
auto_install: false
auto_install_mode: observe_only
```

---

## 🔄 Secure update system

Kyronix Sentinel includes its own GitHub Release update infrastructure.

Check for a newer release:

```bash
sentinelctl update check
```

Example when already current:

```text
Kyronix Sentinel Update

Current: v1.0.0
Latest: v1.0.0
Status: UP TO DATE
```

Install an available update:

```bash
sudo sentinelctl update install
```

---

## 🔁 Update architecture

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
Verify outer SHA256
      │
      ▼
Safe archive extraction
      │
      ▼
Validate manifest
      │
      ▼
Verify every manifest file SHA256
      │
      ▼
Backup current binaries
      │
      ▼
Atomic binary installation
      │
      ▼
Install/update systemd worker unit
      │
      ▼
systemctl daemon-reload
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
      ├── SUCCESS
      │
      └── FAILURE
             │
             ▼
      Restore binaries
             │
             ▼
      Restore worker unit
             │
             ▼
      daemon-reload
             │
             ▼
      Restart previous sentineld
             │
             ▼
      Verify rollback
```

---

## 🔐 Update security

Sentinel `v1.0.0` release archives contain:

```text
manifest.json
sentineld
sentinelctl
sentinel-update.service
```

The manifest defines:

- release version,
- target operating system,
- target architecture,
- expected release files,
- SHA256 hash for every declared file.

Example `v1.0.0` release assets:

```text
sentinel-v1.0.0-linux-arm64.tar.gz
sentinel-v1.0.0-linux-arm64.tar.gz.sha256

sentinel-v1.0.0-linux-amd64.tar.gz
sentinel-v1.0.0-linux-amd64.tar.gz.sha256
```

Unsafe archive contents are rejected, including:

- path traversal,
- absolute paths,
- symbolic links,
- hard links,
- unsupported filesystem objects.

---

## 🛟 Rollback protection

Before replacing the running Sentinel binaries, the updater preserves the previous versions as:

```text
/usr/local/bin/sentineld.previous
/usr/local/bin/sentinelctl.previous
```

The update worker unit is also handled transactionally.

If activation fails, Sentinel restores:

- the previous `sentineld`,
- the previous `sentinelctl`,
- the previous `sentinel-update.service`, or removes the newly provisioned unit if none existed before,
- the previous systemd unit state through `daemon-reload`.

Sentinel then restarts the previous daemon and verifies the previous version through the local health API.

A failed activation is not considered safely rolled back until rollback verification succeeds.

---

## 🚫 Release quarantine

Failed releases that require a verified rollback are automatically quarantined.

This prevents Sentinel from repeatedly attempting to install a release that already failed activation.

Inspect quarantine state:

```bash
sentinelctl update quarantine
```

Clear quarantine manually:

```bash
sudo sentinelctl update quarantine clear
```

---

## 🔒 Update operation lock

Sentinel uses a global update operation lock to prevent concurrent update transactions.

Only one update transaction can modify the installation at a time.

---

## ⚙️ Dedicated update worker

Automatic installation is separated from the long-running Sentinel daemon.

The dedicated worker is:

```text
sentinel-update.service
```

It executes:

```text
/usr/local/bin/sentinelctl update install
```

The worker is a root-owned systemd oneshot service.

A successful execution normally ends as:

```text
LoadState=loaded
ActiveState=inactive
SubState=dead
Result=success
ExecMainStatus=0
```

---

## 🧩 Clean-host bootstrap

`v1.0.0` includes a clean-host and migration bootstrap for the update worker.

If `sentinel-update.service` is missing when the new daemon starts, `sentineld` provisions the canonical worker unit into:

```text
/etc/systemd/system/sentinel-update.service
```

and performs:

```text
systemctl daemon-reload
```

This allows hosts upgrading from earlier development builds to migrate to the `v1.0.0` update architecture without requiring a manually preinstalled worker service.

The bootstrap does not overwrite an existing regular worker unit.

---

## ✅ v1.0.0 release validation

Kyronix Sentinel `v1.0.0` passed the release validation workflow before publication.

Validated areas include:

- complete Go test suite,
- `go vet`,
- release script validation,
- ARM64 release build,
- AMD64 release build,
- external archive SHA256 verification,
- manifest SHA256 verification,
- worker unit inclusion,
- live `v1.0.0` daemon runtime,
- Health Score runtime verification,
- prediction engine runtime verification,
- clean-host worker bootstrap,
- systemd worker loading,
- worker execution,
- worker exit-code contract.

Example validated runtime:

```text
Kyronix Sentinel

Running: true
Health: 100
Risk: LOW
```

Update check:

```text
Kyronix Sentinel Update
Current: v1.0.0
Latest: v1.0.0
Status: UP TO DATE
```

---

## 🐧 Supported builds

| Platform | Architecture | Status |
|---|---|---|
| Linux | ARM64 | ✅ Supported build |
| Linux | AMD64 | ✅ Supported build |

Primary development and validation currently focus on:

- Debian-based Linux
- Ubuntu-based Linux
- Raspberry Pi / ARM64 Linux
- Linux infrastructure hosts
- Kyronix development environments

Planned environment-specific validation includes:

- Proxmox VE hosts
- Docker hosts
- infrastructure appliances
- Kyronix Stratus OS appliances

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

GOTOOLCHAIN=local go build   -o bin/sentineld   ./cmd/sentineld

GOTOOLCHAIN=local go build   -o bin/sentinelctl   ./cmd/sentinelctl
```

---

## 📦 Building a release

Sentinel includes its own release builder.

Example:

```bash
./scripts/build-release.sh 1.0.0
```

Generated artifacts are stored under:

```text
dist/v1.0.0/
```

The release builder automatically:

- builds ARM64 binaries,
- builds AMD64 binaries,
- embeds version metadata,
- embeds Git commit information,
- embeds build date,
- includes the systemd update worker,
- generates release manifests,
- calculates SHA256 hashes for release files,
- creates compressed release archives,
- generates external archive checksums.

---

## ⚙️ systemd

Sentinel runs as a native systemd service.

Main daemon:

```text
sentineld.service
```

Update worker:

```text
sentinel-update.service
```

Check daemon state:

```bash
systemctl status sentineld
```

Quick status:

```bash
systemctl is-active sentineld
```

Check update worker:

```bash
systemctl status sentinel-update.service
```

---

## 🧭 Design principles

### 🧠 Explainable decisions

Every significant prediction should have observable reasons.

### 🧩 Multi-signal consensus

One abnormal metric should not normally trigger a disruptive recovery action.

### 📈 History matters

Persistent degradation is more important than a temporary spike.

### 🛡️ Conservative recovery

Rebooting a production host is disruptive.

Sentinel must be confident before recommending — and eventually performing — recovery.

### 🔐 Safe updates

Software updates must be verified before installation and recoverable when activation fails.

### 🪶 Lightweight by design

Sentinel is not intended to replace:

- Prometheus
- Grafana
- Zabbix
- enterprise observability platforms

Instead, it provides lightweight **local host intelligence** focused specifically on degradation and recovery prediction.

---

## 🗺️ Roadmap

### ✅ Completed for v1.0.0

- Linux host collectors
- CPU collector
- Memory collector
- Disk collector
- Kernel / OOM collector
- Linux PSI support
- Health Score
- Freeze Risk
- Historical trend analysis
- Persistent history
- Multi-signal prediction consensus
- Prediction explainability
- Local Unix socket API
- `sentinelctl`
- GitHub Release discovery
- Architecture-aware update selection
- SHA256 package verification
- Safe archive extraction
- Full manifest file validation
- Atomic binary installation
- Post-install health verification
- Automatic rollback
- Verified rollback tracking
- Failed-release quarantine
- Multi-release quarantine registry
- Persistent update state
- Persistent install lifecycle audit
- `sentinelctl update status`
- `sentinelctl update policy`
- Automatic install policy engine
- Observe-only automatic install evaluation
- Global update operation lock
- Shared update install executor
- Dedicated systemd update worker
- Update worker execution gate
- Worker exit-code contract
- Transactional worker-unit provisioning
- Worker-unit rollback
- systemd `daemon-reload` integration
- Clean-host worker bootstrap
- ARM64 release builds
- AMD64 release builds
- Stable `v1.0.0` release

### 🔜 Next

- Proxmox VE host deployment and validation
- Proxmox-specific health intelligence
- Recovery safety gates
- Controlled automatic recovery execution
- Predictive reboot policy
- Docker host intelligence
- Broader hardware and distribution validation
- Additional prediction heuristics

---

## ⚠️ Production status

Kyronix Sentinel `v1.0.0` is the **first stable release**.

The monitoring, prediction, history, update validation, rollback, quarantine and worker-provisioning infrastructure have reached the first stable release milestone.

However:

- automatic host reboot is not enabled,
- `AUTO_RECOVERY` remains a recommendation,
- automatic update installation is disabled by default,
- the recommended update mode remains conservative and observe-only.

Sentinel should complement — not replace — existing infrastructure monitoring and operational safeguards.

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

https://github.com/FilipooSVK/Kyronix-Sentinel/releases

Current stable release:

```text
v1.0.0
```

Available builds:

```text
Linux ARM64
Linux AMD64
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
