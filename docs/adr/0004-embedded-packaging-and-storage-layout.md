# 0004 - Embedded UI Packaging and Hybrid OS Storage Layout

To ensure seamless installation as an unattended background service while supporting portable evaluation, we decided to embed compiled frontend assets directly into the Go binary via `embed.FS` and adopt a hybrid filesystem storage strategy: by default storing persistent SQLite databases and logs in system directories (`%ProgramData%\Safora` on Windows, `/var/lib/safora` on Linux), with an automatic fallback to local relative paths when `--portable` is flagged.
