# WHOOP TUI

A terminal dashboard for your [WHOOP](https://www.whoop.com) data. View recovery, strain, sleep, and workouts — all from the command line.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the [WHOOP Developer API](https://developer.whoop.com).

## Features

**Dashboard views**
- **Today** — current recovery score, strain, sleep summary, recent workouts
- **Week** — 7-day trends with color-coded bar charts
- **30 / 60 days** — longer-term trends with scrollable day list

**Drill-down detail** for any day:
- Recovery: score, HRV, resting HR, SpO2, skin temp
- Strain: score with bar visualization, avg/max HR, calories
- Sleep: stage breakdown (REM / deep / light / awake), performance, efficiency, consistency, respiratory rate, sleep need analysis
- Workouts: HR zone distribution, distance, elevation

**Color-coded charts** — recovery bars go red-to-green, strain fades light-to-dark blue, sleep shifts red-yellow-green based on duration.

**Responsive** — scales to terminal width and height. Resize and it reflows.

**Cached** — tab switching is instant. Only explicit refresh (`r`) hits the API.

## Prerequisites

- Go 1.21+
- A [WHOOP](https://www.whoop.com) account and strap
- A WHOOP developer app (free to register)

## Setup

### 1. Register a WHOOP developer app

Go to [developer.whoop.com](https://developer.whoop.com) and create an application:

- **Redirect URI**: `http://localhost:8080/callback`
- **Privacy Policy URL**: run `make run` first, then use `http://localhost:8080/privacy` — or host `docs/privacy.html` yourself
- **Scopes**: enable all (`read:profile`, `read:body_measurement`, `read:cycles`, `read:recovery`, `read:sleep`, `read:workout`, `offline`)

### 2. Clone and configure

```bash
git clone https://github.com/Joeavaikath/whoop-tui.git
cd whoop-tui
```

Run the setup command to configure your credentials:

```bash
make setup
```

Or manually create a `.env` file:

```
WHOOP_CLIENT_ID=your-client-id
WHOOP_CLIENT_SECRET=your-client-secret
```

### 3. Build and run

```bash
make build
make run
```

On first run, the app prints an OAuth URL. Open it in your browser, log in to WHOOP, and authorize the app. The token is saved to `~/.config/whoop-tui/token.json` so you won't need to re-authenticate.

## Controls

| Key | Action |
|---|---|
| `1` `2` `3` `4` | Switch to Today / Week / 30d / 60d |
| `Tab` `←` `→` | Cycle between views |
| `↑` `↓` `j` `k` | Scroll day list |
| `Enter` | Drill into selected day |
| `Esc` | Back from detail view |
| `r` | Refresh data from API |
| `q` | Quit |

## Make targets

| Command | Description |
|---|---|
| `make run` | Run the app |
| `make build` | Compile to `bin/whoop-tui` |
| `make install` | Build and copy to `$GOPATH/bin` |
| `make fmt` | Format code |
| `make vet` | Run static analysis |
| `make lint` | Vet + golangci-lint |
| `make test` | Run tests |
| `make deps` | Tidy and download modules |
| `make clean` | Remove build artifacts |

## Project structure

```
cmd/whoop-tui/main.go      Entry point, loads .env, starts auth + TUI
internal/
  auth/oauth.go             OAuth 2.0 flow with local callback server
  api/client.go             Typed WHOOP API client with pagination
  tui/
    app.go                  Bubble Tea program setup
    model.go                State, update logic, tab navigation, caching
    views.go                Dashboard, trend, and detail view rendering
    chart.go                Bar chart renderer with block characters
    styles.go               Shared lipgloss styles and colors
```

## Privacy

This app runs entirely on your machine. No data is sent anywhere except the official WHOOP API. No analytics, no telemetry. OAuth tokens are stored locally in `~/.config/whoop-tui/`.

Full privacy policy: [joeavaikath.github.io/whoop-tui/privacy.html](https://joeavaikath.github.io/whoop-tui/privacy.html)

## License

MIT
