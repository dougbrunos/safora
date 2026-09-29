#!/bin/sh
# Installs Safora as a systemd service. Run as root from the extracted release folder:
#   sudo ./install.sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
  echo "This installer must run as root (use sudo)." >&2
  exit 1
fi

here=$(cd "$(dirname "$0")" && pwd)
bin_dir=/usr/local/bin

# On an upgrade the running service is stopped before its binary is replaced.
[ -x "$bin_dir/safora" ] && "$bin_dir/safora" service stop >/dev/null 2>&1 || true
install -m 0755 "$here/safora" "$bin_dir/safora"
# "install" fails harmlessly on an upgrade because the unit already exists.
"$bin_dir/safora" service install >/dev/null 2>&1 || true
"$bin_dir/safora" service start

echo
echo "Safora $("$bin_dir/safora" version | cut -d' ' -f2) installed and running."
echo "Dashboard: http://127.0.0.1:3434"
echo "Data:      /var/lib/safora (kept on uninstall)"
