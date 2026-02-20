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

# Patch generated package.json: remove prepare hook, pin TypeScript version, set repo/homepage
node -e "
  const fs = require('fs');
  const pkg = JSON.parse(fs.readFileSync('$OUTPUT_DIR/typescript/package.json', 'utf8'));
  delete pkg.scripts.prepare;
  if (pkg.devDependencies) {
    pkg.devDependencies.typescript = '^5.6.0';
  }
  pkg.repository = {
    type: 'git',
    url: 'git+https://github.com/Uttam-Mahata/RootAccess.git'
  };
  pkg.homepage = 'https://github.com/Uttam-Mahata/RootAccess#readme';
  pkg.bugs = {
    url: 'https://github.com/Uttam-Mahata/RootAccess/issues'
  };
  fs.writeFileSync('$OUTPUT_DIR/typescript/package.json', JSON.stringify(pkg, null, 2));
"

# Inject TypeScript client README
cat > "$OUTPUT_DIR/typescript/README.md" << 'TSREADME'
# @rootaccessd/client

TypeScript/Axios API client for any RootAccess-compatible CTF backend, auto-generated
from the OpenAPI spec.

## Installation

```bash
npm install @rootaccessd/client axios
```

## Quick Start

```typescript
import { AuthApi, ChallengesApi, Configuration } from '@rootaccessd/client';

// 1. Point at your backend
const baseConfig = new Configuration({ basePath: 'https://rootaccess.live' });

// 2. Login and extract token
const authApi = new AuthApi(baseConfig);
const { data } = await authApi.authLoginPost({ username: 'you', password: 'secret' });
const token = data.token;

// 3. Use Bearer auth for protected endpoints
const authedConfig = new Configuration({
  basePath: 'https://rootaccess.live',
  apiKey: () => `Bearer ${token}`,
});

const challengesApi = new ChallengesApi(authedConfig);
const challenges = await challengesApi.challengesGet();
```

## Using the RootAccess Angular Frontend

Clone the repo and set your backend URL in
`frontend/src/environments/environment.ts`:

```ts
export const environment = {
  apiUrl: 'https://rootaccess.live',
  wsUrl:  'wss://rootaccess.live',
};
```

Then `npm start` — it talks to your backend automatically.
TSREADME

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

# Patch Python setup.py: set URL and fix description
node -e "
  const fs = require('fs');
  let content = fs.readFileSync('$OUTPUT_DIR/python/setup.py', 'utf8');
  content = content.replace(/url=\"\",/, 'url=\"https://github.com/Uttam-Mahata/RootAccess\",');
  content = content.replace(/long_description=\"\"\"[\\s\\S]*?\"\"\"/, 'long_description=open(\"README.md\").read()');
  fs.writeFileSync('$OUTPUT_DIR/python/setup.py', content);
"

# Inject Python client README
cat > "$OUTPUT_DIR/python/README.md" << 'PYREADME'
# rootaccessd-client

Python API client for any RootAccess-compatible CTF backend, auto-generated from the
OpenAPI spec.

## Installation

```bash
pip install rootaccessd-client
```

## Documentation

Full API documentation and models are available in the [GitHub Repository](https://github.com/Uttam-Mahata/RootAccess/tree/main/clients/python/docs).

- [API Reference](https://github.com/Uttam-Mahata/RootAccess/tree/main/clients/python/docs#documentation-for-api-endpoints)
- [Models Reference](https://github.com/Uttam-Mahata/RootAccess/tree/main/clients/python/docs#documentation-for-models)

## Quick Start

```python
import rootaccessd_client
from rootaccessd_client.api import auth_api, challenges_api

# 1. Point at your backend
config = rootaccessd_client.Configuration(host="https://rootaccess.live")

# 2. Login and extract token
with rootaccessd_client.ApiClient(config) as client:
    api = auth_api.AuthApi(client)
    resp = api.auth_login_post(body={"username": "you", "password": "secret"})
    token = resp["token"]

# 3. Authenticated requests
authed_config = rootaccessd_client.Configuration(
    host="https://rootaccess.live",
    api_key={"Authorization": f"Bearer {token}"},
    api_key_prefix={"Authorization": ""},
)

with rootaccessd_client.ApiClient(authed_config) as client:
    api = challenges_api.ChallengesApi(client)
    challenges = api.challenges_get()
```
PYREADME

echo "✅ All clients generated in the /$OUTPUT_DIR directory!"
