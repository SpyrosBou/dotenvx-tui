#!/bin/bash
set -euo pipefail

VERSION="$1"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 2.1.0"
  exit 1
fi

# Update npm/package.json
node -e "
  const fs = require('fs');
  const p = JSON.parse(fs.readFileSync('npm/package.json', 'utf8'));
  p.version = '$VERSION';
  fs.writeFileSync('npm/package.json', JSON.stringify(p, null, 2) + '\n');
"

echo "Updated npm/package.json to version $VERSION"
