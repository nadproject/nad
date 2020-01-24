#!/usr/bin/env bash
set -ex

dir=$(dirname "${BASH_SOURCE[0]}")
serverDir="$dir/../../.."
outputDir="$serverDir/static"
inputDir="$dir/../../scss"

rm -rf "${outputDir:?}/*"

task="npx node-sass \
  --output-tyle compressed \
  --source-map true \
  --output $outputDir"

# compile first then watch
eval "$task $inputDir"

if [[ $1 == 'true' ]]; then
  eval "$task --watch $inputDir"
fi
