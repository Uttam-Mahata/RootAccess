#!/bin/bash

# .gemini/skills/api-manager/scripts/verify-sdk.sh
CLIENTS_DIR="clients"

echo "🔍 Verifying generated SDKs..."

if [ -d "$CLIENTS_DIR/typescript" ]; then
    TS_VERSION=$(jq -r .version "$CLIENTS_DIR/typescript/package.json")
    echo "✅ TypeScript SDK found (v$TS_VERSION)"
else
    echo "❌ TypeScript SDK missing"
fi

if [ -d "$CLIENTS_DIR/python" ]; then
    echo "✅ Python SDK found"
else
    echo "❌ Python SDK missing"
fi
