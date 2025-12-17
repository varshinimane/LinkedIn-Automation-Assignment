# LinkedIn Automation CLI

CLI-only Go bot that drives LinkedIn via [Rod](https://github.com/go-rod/rod) while applying layered stealth techniques and persisting state in SQLite.

## Project layout
```
linkedin-automation/
├── cmd/               # command dispatcher
├── config/            # config struct + loader
├── auth/              # login & cookie reuse
├── search/            # people search
├── connect/           # connection sending
├── messaging/         # follow-up messages
├── stealth/           # anti-detection utilities
├── storage/           # sqlite + state helpers
├── logger/            # structured logging
├── utils/             # helpers/templating
├── config.yaml.example
├── .env.example
└── main.go
```

## Setup
1. Install Go 1.21+ and Chrome/Chromium.
2. Copy `.env.example` to `.env` and fill `LINKEDIN_EMAIL` / `LINKEDIN_PASSWORD`.
3. Copy `config.yaml.example` to `config.yaml` and tweak limits, templates, filters, and proxy/UA/viewport pool.
4. Install deps (outside this sandbox): `go mod tidy`.

## Run

### Basic Usage
```bash
# Run directly
go run main.go --config=config.yaml

# Or build first
go build -o linkedin-bot .
./linkedin-bot --config=config.yaml start
```

### Usage Examples

**Example 1: Run with default config**
```bash
go run main.go --config=config.yaml
```

**Example 2: Enable debug logging**
Set `debug: true` in `config.yaml` to see detailed debug logs:
```yaml
debug: true
```

**Example 3: Custom search filters**
Edit `config.yaml` to target specific profiles:
```yaml
search:
  job_title: "Software Engineer"
  location: "San Francisco"
  company: "Google"
  pagination_pages: 3
```

**Example 4: Adjust daily limits**
```yaml
limits:
  daily_connections: 20
  daily_messages: 15
```

**Example 5: Use environment variables for credentials**
```bash
export LINKEDIN_EMAIL="your@email.com"
export LINKEDIN_PASSWORD="your_password"
go run main.go --config=config.yaml
```

## Features
- Authentication: Rod login, cookie reuse, detection of bad creds, captcha, 2FA prompt.
- Search: job title/location/company filters, pagination, scroll, dedupe profile URLs.
- Connections: human-like hover/scroll/cursor, personalized notes, daily caps, persistence of contacted profiles.
- Messaging: detect accepted connections, send templated follow-ups, track sent messages.
- Stealth (8+ techniques): human-like mouse with Bezier + micro-corrections/overshoot, randomized timing and think-time, UA pool, viewport variance, navigator.webdriver disable, random scrolling, typing with typos, cursor wandering/hovering, business-hours scheduling, rate-limited cooldowns.
- Persistence: SQLite counters + connection/message history to resume after crash.
- Logging: structured `[INFO]/[WARN]/[ERROR]` output.

## Notes
- Start with low limits and test in headless=false.
- Keep configs/private data out of version control; cookies are stored at `paths.cookie_file` and DB at `paths.db_path`.

