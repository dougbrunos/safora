# Safora

Open-source automated backup management.

Safora replaces fragile, manual backup scripts (such as Windows `.bat` files wrapping `robocopy`) with a resilient, observable background daemon. It is distributed as a single static binary containing the execution engine, an embedded SQLite database, and a React web dashboard.

## Core Features

* **Legacy Script Importer**: Paste existing `robocopy` commands or batch scripts to automatically generate structured Jobs (extracting source paths, destinations, retry policies, and exclusions).
* **Path Template Engine**: Resolve dynamic variables at runtime for integrations with legacy systems (e.g., `D:\Exports\{yesterday:DD-MM-YYYY}`).
* **Windows VSS Integration** (planned): Safely backup open, locked, or active database files using Volume Shadow Copy snapshots.
* **Retention Lock**: Automated pruning of old backups that strictly aborts if the current run fails or if it would delete the last remaining backup copy.
* **Drive Watcher**: Configure jobs to trigger automatically when specific external drives or NAS volumes are mounted.
* **Real-Time Telemetry**: Server-Sent Events (SSE) stream live transfer progress, throughput, and structured logs directly to the embedded web dashboard.

## Installation

Download the latest release from the [Releases page](https://github.com/dougbrunos/safora/releases/latest) and check the file against `SHA256SUMS.txt`.

**Windows**: run `Safora-Setup-<version>.exe` as administrator. It installs Safora as a Windows service and adds a Start menu shortcut to the dashboard. The installer is not code-signed yet, so SmartScreen shows "Unknown publisher": choose *More info*, then *Run anyway*.

**Linux**: extract `safora-<version>-linux-<arch>.tar.gz` and run `sudo ./install.sh`. It installs to `/usr/local/bin` and enables a systemd service. `sudo ./uninstall.sh` removes it and keeps your data (`--purge` deletes it).

The dashboard opens at <http://127.0.0.1:3434> and is only reachable from the same machine. Jobs and history live in `C:\ProgramData\Safora` (Windows) or `/var/lib/safora` (Linux) and survive uninstalling. Start with `--portable` to keep them next to the executable instead.

## Architecture

* **Backend**: Go 1.26. Runs as a background system service (`services.msc` or `systemd`).
* **Database**: Embedded SQLite (WAL mode) for job configurations, run history, and structured logs.
* **Frontend**: React, Vite, Tailwind CSS, and shadcn/ui. The compiled frontend is packaged into the Go executable via `embed.FS` and served on `localhost`.

## Building from Source

### Prerequisites

* Go 1.26+
* Node.js 24+ & npm

### Build Steps

1. Build the frontend assets:
   ```bash
   cd ui/web
   npm install
   npm run build
   ```
2. Build the Go binary:
   ```bash
   cd ../../
   go build -o safora ./cmd/safora
   ```

## Releasing

Releases are built by GitHub Actions (`.github/workflows/release.yml`). To publish one:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds the Linux packages and the Windows installer, computes `SHA256SUMS.txt` and attaches everything to a new GitHub Release. Tags containing a hyphen (for example `v0.2.0-rc1`) are published as pre-releases. To build the packages locally, run `scripts/build-release.sh <version>`; the Windows installer itself needs Inno Setup (see `installer/windows/safora.iss`).

## License

MIT
