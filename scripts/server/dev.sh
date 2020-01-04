#!/usr/bin/env bash
# shellcheck disable=SC1090
# dev.sh builds and starts development environment
set -eux -o pipefail

# clean up background processes
dir=$(dirname "${BASH_SOURCE[0]}")
basePath="$dir/../.."
serverPath=$(realpath "$basePath/pkg/server")
serverPort=3000

# load env
set -a
dotenvPath="$serverPath/.env.dev"
source "$dotenvPath"
set +a

# run server
(cd "$basePath/pkg/watcher" && go run main.go --task="go run main.go start -port $serverPort" --context="$serverPath" "$serverPath")
