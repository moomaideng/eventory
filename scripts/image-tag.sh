#!/usr/bin/env bash
# Prints the content-addressed image tag for one deployable image.
#
#   scripts/image-tag.sh backend <service>   # service under backend/services/
#   scripts/image-tag.sh frontend
#
# The tag is a hash of every tracked file that can change the image, so CI can
# skip a build exactly when an image with that tag is already in the registry.
# Missing an input ships a stale image; including too much only costs a
# rebuild, so every rule here errs towards including more.
#
# Frontend: hashes NEXT_PUBLIC_SUPABASE_URL, NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY
# and NEXT_PUBLIC_API_URL from the environment too, because next build bakes
# them into the image.
set -euo pipefail

# Bump when a build setting outside the hashed files changes the image
# (platforms, build-push-action options in ci.yml), to force a full rebuild.
SCHEME=v1

context=${1:?usage: image-tag.sh backend <service> | image-tag.sh frontend}
service=${2:-}
cd "$(git rev-parse --show-toplevel)"

case "$context" in
  backend)
    : "${service:?image-tag.sh: backend needs a service}"
    # Every in-module package the service's binaries compile, so a shared
    # package (backend/internal, another service's generated proto) changes
    # the tag. Directory pathspecs are recursive, which also picks up
    # embedded and Dockerfile-copied assets such as receipt fonts.
    pkg_dirs=$(cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go list -deps \
      -f '{{with .Module}}{{if .Main}}{{$.Dir}}{{end}}{{end}}' "./services/$service/cmd/...")
    paths=(backend/go.mod backend/go.sum backend/Dockerfile backend/.dockerignore "backend/services/$service")
    while IFS= read -r dir; do
      paths+=("${dir#"$PWD/"}")
    done <<<"$pkg_dirs"
    build_args=""
    ;;
  frontend)
    paths=(frontend)
    build_args=$(printf '%s\n' \
      "${NEXT_PUBLIC_SUPABASE_URL:-}" \
      "${NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY:-}" \
      "${NEXT_PUBLIC_API_URL:-}")
    ;;
  *)
    echo "image-tag.sh: unknown context: $context" >&2
    exit 2
    ;;
esac

# ls-files -s lists mode, blob hash, and path: content and layout, nothing else.
hash=$({
  printf '%s\n%s\n' "$SCHEME" "$build_args"
  git ls-files -s -- "${paths[@]}" | sort -u
} | sha256sum | cut -c1-20)

printf 'src-%s\n' "$hash"
