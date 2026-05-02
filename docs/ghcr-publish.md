# Publishing `auth-provider-ms` to GHCR

This guide explains how to build and push the `auth-provider-ms` container image to the
GitHub Container Registry (GHCR) at `ghcr.io/matiasmartin-labs/auth-provider-ms`.

---

## Prerequisites

1. **Docker** installed and running (Docker Engine ≥ 24 recommended).
2. A **GitHub Personal Access Token (PAT)** with the `write:packages` scope.
   - Create one at <https://github.com/settings/tokens> → *Generate new token (classic)*.
3. Authenticate Docker with GHCR:

```bash
echo "<YOUR_PAT>" | docker login ghcr.io -u <your-github-username> --password-stdin
```

---

## Build

```bash
docker build -t ghcr.io/matiasmartin-labs/auth-provider-ms:latest .
```

To target a specific architecture (e.g., `arm64`):

```bash
docker build --build-arg GOARCH=arm64 \
  -t ghcr.io/matiasmartin-labs/auth-provider-ms:latest .
```

---

## Tag

Replace `<version>` with a semantic version (e.g., `v1.0.0`):

```bash
docker tag ghcr.io/matiasmartin-labs/auth-provider-ms:latest \
           ghcr.io/matiasmartin-labs/auth-provider-ms:<version>
```

---

## Push

```bash
# Push a specific version tag
docker push ghcr.io/matiasmartin-labs/auth-provider-ms:<version>

# Push latest
docker push ghcr.io/matiasmartin-labs/auth-provider-ms:latest
```

---

## Required Environment Variables

The application reads secrets at runtime via Viper environment expansion from `config.yaml`.
Pass the following variables when running the container:

| Variable | Description |
|---|---|
| `GOOGLE_CLIENT_ID` | OAuth2 client ID from Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | OAuth2 client secret from Google Cloud Console |
| `GOOGLE_STATE` | Random string used as the OAuth2 state parameter (CSRF protection) |
| `ALLOWED_EMAILS` | Comma-separated list of e-mail addresses permitted to log in |

---

## Example `docker run`

```bash
docker run --rm -p 8080:8080 \
  -e GOOGLE_CLIENT_ID="your-google-client-id" \
  -e GOOGLE_CLIENT_SECRET="your-google-client-secret" \
  -e GOOGLE_STATE="random-state-string" \
  -e ALLOWED_EMAILS="user@example.com,admin@example.com" \
  ghcr.io/matiasmartin-labs/auth-provider-ms:latest
```

The service will be available at `http://localhost:8080`.

> **Note**: For production deployments set `security.cookie.secure: true` and
> `security.redirect.url` to your actual frontend origin either by mounting a
> custom `config.yaml` or by adding Viper override env vars.
