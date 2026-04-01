# Stockyard Post

**Form backend.** Point any HTML form at it. Collects submissions, fires webhooks, handles redirects. Single binary, no external dependencies.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
curl -sfL https://stockyard.dev/install/post | sh
post

# Or with Docker
docker run -p 8830:8830 -v post-data:/data ghcr.io/stockyard-dev/stockyard-post:latest
```

Dashboard at [http://localhost:8830/ui](http://localhost:8830/ui)

## Usage

```bash
# 1. Create a form
curl -X POST http://localhost:8830/api/forms \
  -H "Content-Type: application/json" \
  -d '{"name":"Contact","redirect_url":"https://mysite.com/thanks"}'
# Returns: {"form": {...}, "submit_url": "http://localhost:8830/f/abc123"}

# 2. Point your HTML form at it
```

```html
<form method="POST" action="http://localhost:8830/f/abc123">
  <input name="email" type="email" required>
  <textarea name="message"></textarea>
  <button type="submit">Send</button>
</form>
```

```bash
# 3. View submissions
curl http://localhost:8830/api/forms/abc123/submissions
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/forms | Create form |
| GET | /api/forms | List forms |
| GET | /api/forms/{id} | Form detail |
| PUT | /api/forms/{id} | Update form |
| DELETE | /api/forms/{id} | Delete form + submissions |
| POST | /f/{id} | Submit form (HTML or JSON) |
| GET | /api/forms/{id}/submissions | List submissions |
| GET | /api/submissions/{id} | Submission detail |
| DELETE | /api/submissions/{id} | Delete submission |
| GET | /api/forms/{id}/export | Export submissions (Pro) |
| GET | /health | Health check |
| GET | /ui | Web dashboard |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8830 | HTTP port |
| DATA_DIR | ./data | SQLite data directory |
| RETENTION_DAYS | 30 | Submission retention |
| POST_LICENSE_KEY | | Pro license key |

## Free vs Pro

| Feature | Free | Pro ($2.99/mo) |
|---------|------|----------------|
| Forms | 3 | Unlimited |
| Submissions/month | 100/form | Unlimited |
| Submission retention | 7 days | 90 days |
| Honeypot spam filter | ✓ | ✓ |
| Webhook notifications | — | ✓ |
| CSV export | — | ✓ |

## License

Apache 2.0 — see [LICENSE](LICENSE).
