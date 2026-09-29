#!/bin/sh
# Removes the Safora service and binary. Backup data and jobs in /var/lib/safora
# are kept unless --purge is given.
set -eu

if [ "$(id -u)" -ne 0 ]; then
  echo "This uninstaller must run as root (use sudo)." >&2
  exit 1
fi

bin=/usr/local/bin/safora
if [ -x "$bin" ]; then
  "$bin" service stop >/dev/null 2>&1 || true
  "$bin" service uninstall >/dev/null 2>&1 || true
  rm -f "$bin"
fi

if [ "${1:-}" = "--purge" ]; then
  rm -rf /var/lib/safora
  echo "Safora removed, including its data."
else
  echo "Safora removed. Jobs and history kept in /var/lib/safora (use --purge to delete)."
fi
