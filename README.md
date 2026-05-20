# StartPlaying.games Clone (Go Version)

A lightweight clone of StartPlaying.games, rewritten in Go for efficiency and containerized for easy deployment.

## Tech Stack

- **Backend:** Go (Echo Framework)
- **Frontend:** HTMX + Tailwind CSS
- **Database:** SQLite
- **Deployment:** Docker / Docker Compose

## Setup & Running

### 1. Locally (with Go installed)

```bash
cd go-backend
go run cmd/server/main.go
```

The app will be available at `http://localhost:30011`.

### 2. With Docker

```bash
docker-compose up --build
```

## Prototype Accounts

For development and testing, you can use the following accounts:

- **GM Account:** `gm` / `password`
- **Player Account:** `player` / `password`

## Features

- **User Roles:** Game Masters (GM) and Players.
- **GM Dashboard:** Create and manage game listings.
- **Player Dashboard:** Join available games and view bookings.
- **Prototype Auth:** Secure password hashing with bcrypt and session management.

## Project Structure

- `go-backend/`: The Go source code, templates, and static assets.
- `Dockerfile`: Multi-stage build for a tiny production image.
- `docker-compose.yml`: Orchestration for app and persistent storage.
