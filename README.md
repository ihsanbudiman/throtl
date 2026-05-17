# Throtl

**Share your AI API access — without sharing your keys.**

Throtl is an API gateway that lets you share your AI subscription — GLM 5.1, Kimi K2.6, or any OpenAI-compatible provider — with your team, friends, or clients, while keeping full control over who uses what, and how much.

> **Note:** Throtl currently supports OpenAI-compatible providers only (any API that follows the `/v1/chat/completions` and `/v1/models` convention).

---

## Why Throtl?

You have an AI subscription — GLM 5.1, Kimi K2.6, or any OpenAI-compatible provider. Your teammate needs access too. You don't want to hand them your API key — they might overspend, use models you don't want, or accidentally leak it.

Throtl sits between your users and the AI provider. You keep your real API key safe. Everyone else gets a share key (`sk-share-...`) that you control.

---

## Features

### 🔑 Share Keys, Not Secrets

Generate share keys for each person or project. Your real provider keys never leave the server. Revoke or disable any key instantly.

### ⏱ Rate Limiting That Makes Sense

Set per-key limits with two independent controls:

- **Window limit** — e.g. 100 requests per 5 hours (the window duration is configurable)
- **Daily limit** — e.g. 500 requests per day, resets at midnight UTC

Both are optional. Set neither for unlimited access. Set one or both to keep spending in check.

### 🧠 Model Control

Connect multiple AI providers and see every available model in one place. Disable any model with a single toggle — disabled models won't appear to users and requests to them are blocked at the gateway.

### 📊 Usage Dashboard

See who's using what, how many tokens are flowing, and which models are popular — all in real time.

### 🌙 Dark Mode

Because staring at a white dashboard at 2 AM hurts.

### 🔒 Single Admin, Zero Complexity

One admin account. No role management, no team invites, no SSO. Simple.

---

## How It Works

```
Your User → sk-share-... → Throtl Gateway → Your Provider API Key → OpenAI-compatible Provider
                                │
                                ├── Rate Limiter (window + daily)
                                ├── Model Access Control
                                ├── Usage Logger
                                └── Request Proxy
```

1. You add your provider (e.g. any OpenAI-compatible API) with its real API key
2. You create a share key with limits and allowed models
3. Your user calls `https://your-throtl/v1/chat/completions` with their share key
4. Throtl checks limits, verifies the model is allowed, proxies the request, and logs usage

Users call the same OpenAI-compatible endpoints they're used to — just with a different base URL and key.

---

## Quick Start

### Requirements

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### Run

```bash
git clone https://github.com/ihsanbudiman/throtl.git
cd throtl
docker compose up -d
```

That's it. Open **http://localhost:3000** and create your admin account.

### First-Time Setup

1. You'll see the setup page — create your admin email and password
2. Add a **Provider** — any OpenAI-compatible API (e.g. `https://api.openai.com/v1`, or your self-hosted endpoint) with your API key
3. Create an **API Key** — set rate limits and allowed models
4. Share the generated `sk-share-...` key with your user

### Using a Share Key

Replace the base URL and API key in any OpenAI-compatible client:

```bash
curl https://your-throtl/v1/chat/completions \
  -H "Authorization: Bearer sk-share-..." \
  -H "Content-Type: application/json" \
  -d '{
    "model": "provider-id/GLM-5.1",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

The model name uses the format `provider-id/model-name` (e.g. `wafer/GLM-5.1`). Users can discover available models:

```bash
curl https://your-throtl/v1/models \
  -H "Authorization: Bearer sk-share-..."
```

---

## Ports

| Service | Port |
|---------|------|
| Dashboard | http://localhost:3000 |
| Gateway API | http://localhost:8081 |
| PostgreSQL | localhost:5432 |

---

## Configuration

Set these environment variables in `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| `THROTL_PORT` | `8080` | Gateway server port (inside container) |
| `THROTL_DB_URL` | `postgres://throtl:throtl@db:5432/throtl` | PostgreSQL connection string |
| `THROTL_JWT_SECRET` | auto-generated | JWT signing key. **Set this to a secure random string** — otherwise sessions invalidate on every restart |

---

## Stopping

```bash
docker compose down
```

Data persists in the `pgdata` Docker volume. To wipe everything:

```bash
docker compose down -v
```

---

## Tech Stack

| Layer | Tech |
|-------|------|
| Backend | Go + Echo |
| Frontend | React + Vite + Tailwind CSS |
| Database | PostgreSQL 17 |
| Auth | JWT + bcrypt |

---

## License

MIT — do whatever you want, just keep the license notice.
