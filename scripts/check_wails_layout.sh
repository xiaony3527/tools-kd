#!/usr/bin/env bash
set -euo pipefail

errors=0

check_file() {
    if [ ! -f "$1" ]; then
        echo "MISSING: $1"
        errors=$((errors + 1))
    else
        echo "  FOUND: $1"
    fi
}

echo "Checking Wails project layout..."
echo ""

check_file "wails.json"
check_file "frontend/package.json"
check_file "frontend/tsconfig.json"
check_file "frontend/vite.config.ts"
check_file "frontend/index.html"
check_file "frontend/src/main.ts"
check_file "frontend/src/App.ts"
check_file "frontend/src/styles.css"

echo ""
if [ "$errors" -eq 0 ]; then
    echo "All layout checks PASSED."
    exit 0
else
    echo "Layout checks FAILED: $errors file(s) missing."
    exit 1
fi
