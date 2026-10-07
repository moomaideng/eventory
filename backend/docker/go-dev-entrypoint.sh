#!/bin/sh
set -eu

: "${SERVICE:?SERVICE is required}"

# Shared .air.toml uses $SERVICE; air does not expand env vars itself.
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
sed "s|\$SERVICE|${SERVICE}|g" /app/.air.toml > "$tmpdir/air.toml"

exec air -c "$tmpdir/air.toml"
