# Unkey Deploy demo

A small Go API that uses [Zen](https://github.com/unkeyed/unkey/tree/main/pkg/zen)
to demonstrate Unkey Deploy and gateway API key authentication. It needs no
database, root key, or runtime configuration. `PORT` defaults to 8080.

| Route | Response |
| --- | --- |
| `GET /` | Gateway setup instructions |
| `GET /health` | `{"status":"ok"}` |
| `GET /api/hello` | A greeting and request ID |

The container is public by default. Unkey's gateway verifies API keys, not this
app. The app does not trust principal headers or echo credentials. Keep the
container behind the gateway after you configure authentication.

## Run and test

Use Go 1.27.2 and [mise](https://mise.jdx.dev/):

```bash
mise run test
mise run run
```

In another terminal:

```bash
curl http://localhost:8080/api/hello
```

Zen is pinned to an Unkey commit in `go.mod`. This repository does not need a
local Unkey checkout and does not copy or fork the framework.
Zen is distributed under [Unkey's license](https://github.com/unkeyed/unkey/blob/c60bb4657b3eb217b7c7703b68078fed9616a924/LICENSE).

## Build a container

With Docker and its multi-platform image store enabled:

```bash
mise run image
docker run --rm --read-only --cap-drop=ALL -p 8080:8080 deploy-demo:local
```

This builds static binaries for Linux amd64 and arm64 and packages them in a
non-root distroless image. `mise run build` produces the binaries without Docker.
The Dockerfile expects these binaries to exist in `dist/`.

## Protect the API

After the image is published and public:

1. Deploy `ghcr.io/unkeyed/deploy-demo:v0.1.0` as a container image in Unkey.
   Set the runtime port to 8080 and the health check path to `/health`.
2. Create a keyspace and an API key in **API Management**, in the same workspace.
3. Open the app's **Policies**, add **Key Auth**, and select your keyspace.
   Match the path prefix `/api/`, use Bearer authentication, and set credit cost
   to 0. Leave `/` and `/health` public.
4. Redeploy to activate the policy.
5. Request `/api/hello` without a key. Expect 401.
6. Repeat with the API key. Expect 200 and a greeting.
7. Revoke the key and retry. Expect 401.

```bash
export DEMO_URL='https://your-deployment-domain'
curl -i "$DEMO_URL/api/hello"
read -rs -p 'Demo API key: ' DEMO_API_KEY; echo
curl -i "$DEMO_URL/api/hello" \
  -H "Authorization: Bearer $DEMO_API_KEY"
unset DEMO_API_KEY
```

Use an end-user API key, never a root key. Add permissions or per-key rate limits
to try 403 and 429 responses. See
[gateway API key authentication](https://www.unkey.com/docs/compute/gateway/api-key-auth).

## Publish

The GitHub Actions workflow tests pushes to `main`, pull requests, and version
tags. A tag such as `v0.1.0` publishes both image architectures to
`ghcr.io/unkeyed/deploy-demo:v0.1.0`, using the workflow's `GITHUB_TOKEN`.

After the first publication, an organization package admin must change the GHCR
package visibility to public. Public source does not make the image public.
Verify an anonymous pull before using the image as a dashboard placeholder.
