#!/usr/bin/env bash
set -eux

dir=$(dirname "${BASH_SOURCE[0]}")
"$dir/build.sh" true
