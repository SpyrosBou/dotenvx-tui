#!/bin/bash
set -euo pipefail

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 2.1.0"
  exit 1
fi

# Update npm/package.json
VERSION="$VERSION" node <<'NODE'
  const fs = require('fs');
  const version = process.env.VERSION;

  if (!/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/.test(version)) {
    console.error('version must be a valid semver string without a leading v');
    process.exit(1);
  }

  const p = JSON.parse(fs.readFileSync('npm/package.json', 'utf8'));
  p.version = version;
  fs.writeFileSync('npm/package.json', JSON.stringify(p, null, 2) + '\n');
NODE

echo "Updated npm/package.json to version $VERSION"
