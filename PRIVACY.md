# Privacy Policy — WHOOP TUI

**Last updated:** April 23, 2026

## Overview

WHOOP TUI is a personal, open-source terminal application that displays your WHOOP data locally on your computer. It is not a hosted service.

## Data Collection

WHOOP TUI does **not** collect, store, transmit, or share any personal data with third parties. All data retrieved from the WHOOP API is displayed locally in your terminal and stored only on your machine.

## Data Storage

- **OAuth tokens** are stored locally in `~/.config/whoop-tui/` on your machine.
- **No data is sent** to any server other than the official WHOOP API (`api.prod.whoop.com`).
- **No analytics, telemetry, or tracking** of any kind is used.

## WHOOP API Access

This application uses the WHOOP Developer API with your explicit authorization via OAuth 2.0. You can revoke access at any time through your WHOOP account settings.

## Scopes Requested

- `read:profile` — Your name and email
- `read:body_measurement` — Height, weight, max heart rate
- `read:cycles` — Daily strain and cycle data
- `read:recovery` — Recovery scores, HRV, resting heart rate
- `read:sleep` — Sleep stages and performance
- `read:workout` — Workout activity data
- `offline` — Refresh tokens for persistent login

## Contact

If you have questions about this privacy policy, open an issue on the project's GitHub repository.
