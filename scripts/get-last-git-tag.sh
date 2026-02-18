#!/usr/bin/env sh
set -eu

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "Error: not inside a Git repository." >&2
  exit 1
fi

# Keep local tags in sync with remote before selecting the latest one.
# If network/remote access is unavailable, continue with local tags.
if ! git fetch --tags --quiet 2>/dev/null; then
  echo "Warning: could not fetch tags from remote; using local tags." >&2
fi

last_tag="$(git tag --sort=-creatordate | head -n 1)"

if [ -z "$last_tag" ]; then
  echo "Error: no Git tags found." >&2
  exit 1
fi

printf '%s\n' "$last_tag"
