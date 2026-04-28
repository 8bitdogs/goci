# GOCI

[![Docker Image Version](https://img.shields.io/docker/v/8bitdogs/goci?sort=semver&label=latest)](https://hub.docker.com/r/8bitdogs/goci)
[![Docker Hub Pulls](https://img.shields.io/docker/pulls/8bitdogs/goci)](https://hub.docker.com/r/8bitdogs/goci)
[![Docker Stars](https://img.shields.io/docker/stars/8bitdogs/goci)](https://hub.docker.com/r/8bitdogs/goci)
[![Docker Cloud Automated build](https://img.shields.io/docker/cloud/build/8bitdogs/goci.svg)](https://hub.docker.com/r/8bitdogs/goci)
[![GitHub](https://img.shields.io/github/license/8bitdogs/goci)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/8bitdogs/goci)](https://goreportcard.com/report/github.com/8bitdogs/goci)

## Description

**goci** is a lightweight CI/CD runner that deploys your services **without SSH keys**. It listens for GitHub webhook events, runs a configurable pipeline (pull image, redeploy container, run commands), and reports the result back to GitHub as a commit status. Your GitHub Actions workflow waits for the status and finishes accordingly — giving you a clean, auditable deployment gate directly inside your PR/push workflow.

Key features:
- No SSH keys or VPN required on the target host
- Driven by GitHub webhooks (`workflow_job` or `push` events)
- Commit status integration — GitHub Actions waits for goci to finish
- Per-service pipeline configuration (JSON or YAML)
- Docker Compose-aware: pulls new images and redeploys services

---

## How it works

```
           GITHUB                                       GOCI
──────────────────────────────────────────────────────────────────────
  git push
       │
       ▼
  Actions: build/push job
  ├─ build container image
  └─ push image to registry
       │
       │  workflow_job.completed
       │ ──────────────────────────────────────────▶ receive webhook
       │                                                   │
       │                                            [1] set commit status: pending
       │                                                   │
  Actions: deploy job                               [2] run pipeline
  (waiting for commit status)                              ├─ pull new image
       │                                                   ├─ redeploy service
       │                                                   └─ run steps
       │                                                    │
       │                                            [3] set commit status
       │                                             ┌──────┴────────┐
       │                                          success         failure
       │                                             │               │
       │◀───────────────── commit status: success ──┘               │
       │◀───────────────────────────── commit status: failure ──────┘
       │
       ▼
  deploy job: ✓ success / ✗ failure
```

---

## Configuration

### Environment variables (`.env` / top-level)

These are global defaults. Any value can be overridden per-service in the config file.

| Variable | Default | Description |
|---|---|---|
| `CI_HOST` | _(empty)_ | Optional. Public URL of this goci instance. Used in status update descriptions. |
| `SERVER_ADDR` | `:7878` | Address and port for the goci HTTP server. |
| `LOG_LEVEL` | `info` | Log verbosity: `debug`, `info`, `warn`, `error`. |
| `GITHUB_TOKEN` | _(required)_ | GitHub personal access token (or `GITHUB_TOKEN` from Actions). Used to set commit statuses. |
| `GITHUB_WEBHOOK_SECRET` | _(empty)_ | Secret used to validate incoming GitHub webhooks. |
| `GITHUB_METHOD` | `POST` | HTTP method expected for the webhook endpoint. |
| `GITHUB_RESPONSE_TIMEOUT` | `10s` | Timeout for outbound GitHub API calls. |
| `GITHUB_TARGET_BRANCH` | `main` | Only process events targeting this branch. |
| `GITHUB_EVENT_TYPE` | `push` | GitHub event type to listen for. Use `workflow_job` for workflow-driven deployments. |
| `GITHUB_COMMIT_STATUS_CONTEXT` | `goci/deploy` | Context string for GitHub commit status updates. Must match the `context` field in your GitHub Actions `wait-commit-status` step. |
| `GITHUB_WORKFLOW_NAME` | _(empty)_ | Optional. Filter by workflow name (the `name:` field at the root of the workflow YAML). |
| `GITHUB_WORKFLOW_JOB_NAME` | _(required for `workflow_job`)_ | Name of the GitHub Actions job whose completion triggers goci. |
| `GITHUB_WORKFLOW_ACTION` | `completed` | Workflow job action to react to: `queued`, `in_progress`, or `completed`. |

See [`.env.example`](.env.example) for a ready-to-copy template.

---

### Service config file (`config.yaml` / `config.json`)

Pass the path with `-config /path/to/config.yaml` (default: `config.json`).

Defines one or more services, each with its own pipeline and GitHub webhook settings. Service-level GitHub fields override the global environment variables.

**YAML example:**

```yaml
- name: "my-service"
  pipeline:
    jobs:
      - name: "deploy"
        steps:
          - name: "pull image"
            cmd: "docker"
            args: ["compose", "pull", "my-service"]
            dir: "/app"
            timeout: 60s
          - name: "restart service"
            cmd: "docker"
            args: ["compose", "up", "-d", "my-service"]
            dir: "/app"
            timeout: 30s
  github:
    path: "/webhook/my-service"          # GitHub webhook URL path
    method: "POST"                        # optional, default POST
    secret: "<webhook-secret>"            # overrides GITHUB_WEBHOOK_SECRET
    token: "<github-token>"              # overrides GITHUB_TOKEN
    branch: "main"                        # target branch filter
    event: "workflow_job"                 # event type
    commit_status_context: "goci/deploy" # overrides GITHUB_COMMIT_STATUS_CONTEXT
    workflow:
      action: "completed"
      job_name: "build-and-push"         # GitHub Actions job name that triggers goci
```

**JSON example:** see [`config.json.example`](config.json.example).

---

## Pull the image

```sh
# Docker Hub
docker pull 8bitdogs/goci:latest

# GitHub Container Registry
docker pull ghcr.io/8bitdogs/goci:latest
```

---

## Docker Compose setup

Add goci as a service in your `docker-compose.yaml`:

```yaml
name: "<compose-name>"  # required

services:

  # your application services...

  goci:
    image: 8bitdogs/goci:latest
    container_name: goci
    depends_on:
      - your-service
    env_file:
      - .env                                          # top-level configuration
    volumes:
      - ~/.docker/config.json:/root/.docker/config.json:ro  # registry auth
      - /var/run/docker.sock:/var/run/docker.sock           # Docker control
      - ./docker-compose.yaml:/docker-compose.yaml          # Compose file for redeployment
      - ./config.yaml:/config.yaml                          # goci service config
    command: goci -config /config.yaml
```

> **Note:** Mounting `docker.sock` gives goci the ability to pull images and restart containers on the host. Ensure the goci container is trusted.

---

## GitHub Actions integration

Add a `deploy` job to your workflow that waits for goci to report the deployment result via commit status:

```yaml
jobs:
  docker:
    name: build and push docker image
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build and push
        # ... your build/push steps

  deploy:
    name: deploy
    needs: docker
    runs-on: ubuntu-latest
    steps:
      - uses: 8bitdogs/wait-commit-status@main
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          context: goci/deploy   # must match GITHUB_COMMIT_STATUS_CONTEXT or service config
          interval: 5            # poll every 5 seconds
          timeout: 300           # fail after 5 minutes
```

The `deploy` job will block until goci sets the commit status to `success` or `failure`, ensuring your workflow reflects the actual deployment result.

> The `context` value must exactly match `GITHUB_COMMIT_STATUS_CONTEXT` (env) or `commit_status_context` in your service config.

---

## GitHub Webhook setup

Go to your repository (or organization) → **Settings → Webhooks → Add webhook**.

### 1. Payload URL

Combine your goci host with the webhook path defined in your service config:

```
<CI_HOST> + <github.path from service config>

# example
https://goci.example.com/webhook/my-service
```

### 2. Content type

Select **`application/json`**.

### 3. Secret

Generate a high-entropy secret and paste it into the **Secret** field. Use the same value for `GITHUB_WEBHOOK_SECRET` (env) or `secret` in your service config.

```sh
openssl rand -hex 32
```

### 4. Events

Choose **"Let me select individual events"** and enable:

| Event | Purpose |
|---|---|
| **Ping** | Verifies the webhook is reachable (always enable) |
| **Push** | Triggers goci on `push` events (`GITHUB_EVENT_TYPE=push`) |
| **Workflow jobs** | Triggers goci when a workflow job completes (`GITHUB_EVENT_TYPE=workflow_job`) |

Disable all other events to reduce noise.

### 5. Verify

After saving, GitHub immediately sends a `ping` event. Open the **Recent Deliveries** tab and check the response code:

| Status | Meaning |
|---|---|
| `200` | Webhook validated and pipeline executed |
| `202` | Webhook validated but nothing matched (no pipeline executed) |
| `4xx` / `5xx` | Error — see the response payload message for details |

Any other status or a failed delivery means goci is unreachable or the payload was rejected — check that goci is running, the Payload URL is correct, and the secret matches.

---

## Quick start

```sh
# 1. Copy and fill in environment config
cp .env.example .env

# 2. Create your service config
cp config.yaml.example config.yaml

# 3. Run with Docker Compose
docker compose up -d goci
```

Or build from source:

```sh
make up
```
