#!/bin/bash
set -e

# Configuration
SPEC_FILE="backend/docs/swagger.json"
OUTPUT_DIR="clients"

# Check if swagger file exists
if [ ! -f "$SPEC_FILE" ]; then
    echo "Error: Swagger specification not found at $SPEC_FILE"
    echo "Run 'cd backend && ~/go/bin/swag init -g cmd/api/main.go' first."
    exit 1
fi

echo "🚀 Generating API Clients..."

# Generate TypeScript (Axios)
echo "📦 Generating TypeScript (Axios) client..."
rm -rf "$OUTPUT_DIR/typescript"
mkdir -p "$OUTPUT_DIR/typescript"
npx @openapitools/openapi-generator-cli generate \
    -i "$SPEC_FILE" \
    -g typescript-axios \
    -o "$OUTPUT_DIR/typescript" \
    --skip-validate-spec \
    --additional-properties=npmName=@rootaccessd/client,supportsES6=true

# Generate Python
echo "📦 Generating Python client..."
rm -rf "$OUTPUT_DIR/python"
mkdir -p "$OUTPUT_DIR/python"
npx @openapitools/openapi-generator-cli generate \
    -i "$SPEC_FILE" \
    -g python \
    -o "$OUTPUT_DIR/python" \
    --skip-validate-spec \
    --additional-properties=packageName=rootaccessd_client,projectName=rootaccessd-client

echo "✅ All clients generated in the /$OUTPUT_DIR directory!"
