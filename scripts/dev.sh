#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

for env_file in backend/.env frontend/.env.local; do
  if [[ ! -f "$env_file" ]]; then
    example_file="${env_file%.local}.example"
    if [[ "$env_file" == "backend/.env" ]]; then
      example_file="backend/.env.example"
    fi
    echo "Missing $env_file. Copy $example_file to $env_file and configure it." >&2
    exit 1
  fi
done

for command_name in docker go npm; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command not found: $command_name" >&2
    exit 1
  fi
done

air_command="$(command -v air || true)"
if [[ -z "$air_command" ]]; then
  go_bin="$(go env GOBIN)"
  if [[ -z "$go_bin" ]]; then
    go_bin="$(go env GOPATH)/bin"
  fi
  air_command="$go_bin/air"
fi
if [[ ! -x "$air_command" ]]; then
  echo "Air is not installed. Run 'make install' first." >&2
  exit 1
fi

compose=(docker compose --env-file backend/.env --env-file frontend/.env.local)

echo "Stopping containerized frontend and backend..."
"${compose[@]}" stop frontend backend

echo "Starting PostgreSQL..."
"${compose[@]}" up -d --wait postgres

echo "Running database migrations..."
(
  cd backend
  go run ./cmd/migrate
)

backend_pid=""
frontend_pid=""

cleanup() {
  local exit_code=$?
  trap - EXIT INT TERM

  if [[ -n "$backend_pid" ]]; then
    kill "$backend_pid" 2>/dev/null || true
  fi
  if [[ -n "$frontend_pid" ]]; then
    kill "$frontend_pid" 2>/dev/null || true
  fi

  wait "$backend_pid" "$frontend_pid" 2>/dev/null || true
  exit "$exit_code"
}
trap cleanup EXIT INT TERM

echo "Starting backend with Air and frontend with Next.js..."
(
  cd backend
  exec "$air_command" -c .air.toml
) &
backend_pid=$!

(
  cd frontend
  exec npm run dev
) &
frontend_pid=$!

set +e
wait -n "$backend_pid" "$frontend_pid"
status=$?
set -e
exit "$status"
