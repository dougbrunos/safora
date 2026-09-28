# 0002 - OS Daemon Service with Local Web UI and Desktop Toasts

To ensure backups execute reliably without an active logged-in user while retaining seamless desktop usability, we decided to run the core Safora engine as an operating system service (Windows Service / Linux systemd daemon) serving an embedded local web dashboard, with a lightweight desktop companion for Windows Toast notifications. This avoids the fragility and high memory overhead of full desktop runtimes (like Electron) while guaranteeing scheduled and device-arrival backups run uninterrupted.
