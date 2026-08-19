#!/usr/bin/env bash
# Intended for a Debian 13 development host with internet access.
# This sandbox cannot execute the downloads, so the script is committed but unexecuted here.
set -euo pipefail

GO_VERSION=1.26.6
GO_SHA256=708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89
NODE_VERSION=24.19.0
NODE_SHA256=14b342e71204f811bde6153be8e04b62aef63c236fef92b55f9c83154b409647
PNPM_VERSION=10.34.5

workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

curl -fL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o go.tar.gz
echo "${GO_SHA256}  go.tar.gz" | sha256sum -c -
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go.tar.gz

curl -fL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" -o node.tar.xz
echo "${NODE_SHA256}  node.tar.xz" | sha256sum -c -
sudo rm -rf "/usr/local/node-v${NODE_VERSION}"
sudo mkdir -p "/usr/local/node-v${NODE_VERSION}"
sudo tar -C "/usr/local/node-v${NODE_VERSION}" --strip-components=1 -xf node.tar.xz
for bin in node npm npx corepack; do
  sudo ln -sf "/usr/local/node-v${NODE_VERSION}/bin/${bin}" "/usr/local/bin/${bin}"
done

sudo /usr/local/bin/npm install -g "pnpm@${PNPM_VERSION}"

printf 'Installed:\n'
/usr/local/go/bin/go version
/usr/local/bin/node --version
/usr/local/bin/pnpm --version

cat <<'MSG'

Docker/Compose should be installed using Docker's current official Debian instructions,
then run ./scripts/check-toolchain.sh before continuing.
MSG
