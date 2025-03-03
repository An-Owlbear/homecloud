#!/usr/bin/env bash

INPUT_DIR="packages";
cd $INPUT_DIR || exit;

for file in ./*.json; do
  if [ -f "$file" ]; then
    echo "Migrating $file";
    appName=$(basename "$file" .json);
    mkdir -p "$appName";
    mv "$file" "$appName/schema.json";
  fi
done