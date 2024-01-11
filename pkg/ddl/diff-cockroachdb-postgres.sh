#!/usr/bin/env bash
set -Eeuo pipefail

# https://github.com/ginokent/cdiff/blob/cbf77fa4186b309c829be3b15fa00b99e563de7c/bin/cdiff#L36
cdiff() { (
  if command -v diff-so-fancy >/dev/null; then
    diff -u "$@" | diff-so-fancy
  else
    if [ -t 0 ]; then
      P=printf C="\033" R=$($P "$C\[31m")
      G=$($P "$C\[32m")
      B=$($P "$C\[36m")
      W=$($P "$C\[1m")
      N=$($P "$C\[0m")
    fi
    diff -u "$@" | sed "s/^\(@@..*@@\)$/${B-}\1${N-}/;s/^\(+.*\)/${G-}\1${N-}/;s/^\(-.*\)/${R-}\1${N-}/;s/^${G-}\(+++ [^ ].*\)/${W-}\1/;s/^${R-}\(--- [^ ].*\)/${W-}\1/;"
  fi
); }
export -f cdiff

diff_envs() { cdiff "$@" | perl -pe "s/(Only in .*: .*)/\033\[1;33m\1\033\[0m/"; }
export -f diff_envs

cd "$(dirname "$0")"

diff_envs \
  --recursive \
  --exclude="*_test.go" \
  --ignore-blank-lines \
  --ignore-space-change \
  --ignore-matching-lines="//diff:ignore-line-postgres-cockroach" \
  --ignore-matching-lines="package postgres" \
  --ignore-matching-lines="package cockroachdb" \
  cockroachdb \
  postgres |
  less --tabs=4 -RFX
