#!/usr/bin/env bash

INPUT_DIR="packages";
cd $INPUT_DIR || exit;

for packageDir in ./*; do
  if [ -d "$packageDir" ]; then
    echo "Creating bundle for $packageDir";
    touch  "$packageDir"/package.tar.gz;
    (cd "$packageDir" && gtar -czf "package.tar.gz" --exclude=package.tar.gz . --owner=0 --group=0 --no-same-owner --no-same-permissions);
  fi
done