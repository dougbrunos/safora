#!/bin/sh
# Builds the frontend and packages Safora for distribution.
#   scripts/build-release.sh [version]
# Output in dist/:
#   safora-<version>-linux-amd64.tar.gz, safora-<version>-linux-arm64.tar.gz  (binary + install.sh + uninstall.sh)
#   safora-<version>-windows-amd64.zip                                        (portable executable)
#   windows-amd64/safora.exe                                                  (input for the Inno Setup installer)
set -eu

version=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"

echo "==> Frontend"
(cd ui/web && npm ci && npm run build)

rm -rf dist
mkdir -p dist

build() { # os arch
  out="dist/$1-$2"
  mkdir -p "$out"
  bin=safora
  [ "$1" = windows ] && bin=safora.exe
  echo "==> $1/$2"
  CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -trimpath \
    -ldflags "-s -w -X main.version=$version" -o "$out/$bin" ./cmd/safora
}

for arch in amd64 arm64; do
  build linux "$arch"
  cp installer/linux/install.sh installer/linux/uninstall.sh "dist/linux-$arch/"
  # The packaged scripts must be executable whatever mode Git checked them out with.
  chmod 0755 "dist/linux-$arch/install.sh" "dist/linux-$arch/uninstall.sh" "dist/linux-$arch/safora"
  tar -C "dist/linux-$arch" -czf "dist/safora-$version-linux-$arch.tar.gz" .
done

build windows amd64
if command -v zip >/dev/null 2>&1; then
  (cd dist/windows-amd64 && zip -q "../safora-$version-windows-amd64.zip" safora.exe)
fi

echo
echo "Done: $version"
ls -1 dist
echo
echo "Windows installer: on a Windows machine with Inno Setup run"
echo "  iscc /DVersion=$version installer\\windows\\safora.iss"
