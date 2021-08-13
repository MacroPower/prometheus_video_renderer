#!/bin/bash

#
# Usage:
#   scripts/load.sh bad_apple
#

set -eu

source .bingo/variables.env

project=$1

if [[ ! -d metrics/$project ]]; then
  echo "No project found in metrics/$project"
fi

for file in metrics/$project/*; do
  filename=$(basename $file)
  echo "Loading $filename"
  $PROMTOOL tsdb create-blocks-from openmetrics "metrics/$project/$filename"
done
