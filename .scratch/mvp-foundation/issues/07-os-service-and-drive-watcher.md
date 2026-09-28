# 07: OS Service Daemon, Drive Arrival Watcher & Desktop Toasts

**What to build:**
The background automation and operating system integration layer. It allows Safora to install and run as a Windows Service (`services.msc`) or Linux systemd daemon to execute scheduled jobs unattended, watches for external volume connections (automatically triggering associated Jobs when a USB drive or target drive like `F:` is connected), and triggers native Windows Desktop Toast notifications when backups complete or fail.

**Blocked by:** 05 (REST API & Real-time SSE Telemetry Streaming), 06 (Embedded Web Dashboard & Script Importer UI)

**Status:** ready-for-agent

- [ ] OS Service manager supporting `safora service [install | uninstall | start | stop | status]` using `kardianos/service` (standard Go cross-platform service library).
- [ ] In-process scheduler supporting cron expressions and interval-based Job triggering.
- [ ] Drive Arrival Watcher monitoring volume mount events and drive letters (e.g., when drive `F:` arrives) to automatically trigger linked Jobs.
- [ ] Desktop notifier displaying native Windows toast notifications (Success / Failure with technical reason) upon Run completion.
- [ ] End-to-end integration test verifying scheduled execution and notification dispatch.
