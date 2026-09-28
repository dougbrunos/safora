# Safora

Open-source automated backup management.

Safora replaces fragile, manual backup scripts (such as Windows `.bat` files wrapping `robocopy`) with a resilient, observable background daemon. It is distributed as a single static binary containing the execution engine, an embedded SQLite database, and a React web dashboard.

## Core Features

* **Legacy Script Importer**: Paste existing `robocopy` commands or batch scripts to automatically generate structured Jobs (extracting source paths, destinations, retry policies, and exclusions).
* **Path Template Engine**: Resolve dynamic variables at runtime for integrations with legacy systems (e.g., `D:\Exports\{yesterday:DD-MM-YYYY}`).
* **Windows VSS Integration**: Safely backup open, locked, or active database files using Volume Shadow Copy snapshots.
* **Retention Lock**: Automated pruning of old backups that strictly aborts if the current run fails or if it would delete the last remaining backup copy.
* **Drive Watcher**: Configure jobs to trigger automatically when specific external drives or NAS volumes are mounted.
* **Real-Time Telemetry**: Server-Sent Events (SSE) stream live transfer progress, throughput, and structured logs directly to the embedded web dashboard.

## Architecture

* **Backend**: Go 1.24. Runs as a background system service (`services.msc` or `systemd`).
* **Database**: Embedded SQLite (WAL mode) for job configurations, run history, and structured logs.
* **Frontend**: React, Vite, Tailwind CSS, and shadcn/ui. The compiled frontend is packaged into the Go executable via `embed.FS` and served on `localhost`.

## Building from Source

### Prerequisites

* Go 1.24+
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


## License

MIT
