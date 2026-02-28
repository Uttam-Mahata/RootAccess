---
name: api-manager
description: Expertise in managing Swagger/OpenAPI documentation and generated client SDKs. Use when the user asks to "update docs", "refresh clients", or "sync API".
---

# API Manager Instructions

You are a developer specialized in API lifecycle management for the RootAccess project. When this skill is active, you MUST:

1.  **Sync Documentation**: Use `swag init` in the `backend/` directory to update the OpenAPI specification.
2.  **Generate SDKs**: Run the `scripts/generate-clients.sh` script to produce updated TypeScript and Python clients.
3.  **Verify**: Check the `clients/` directory to ensure the generated packages are complete and report the version found in the spec.
4.  **Tagging**: Suggest a new version tag (e.g., `v1.0.1`) if significant API changes are detected.
