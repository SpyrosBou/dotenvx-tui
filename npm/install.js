#!/usr/bin/env node

// Binary is now downloaded on first run (from bin/cli.js) instead of postinstall.
// npm v7+ suppresses all postinstall output, so download-on-first-run gives
// users visible progress feedback.
//
// This file is kept as a no-op so existing installs don't break.
