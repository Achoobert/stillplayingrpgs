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

### 3. Startup & Architecture Description

The application processes requests as a monolithic Go service utilizing the Echo framework. On startup, it:

1. Connects to or creates the SQLite database specified by `DATABASE_URL` (defaults to `./database.db`).
2. Configures the db schema (creates `users`, `games`, `bookings` tables).
3. If no users exist, seeds prototype data: a GM (`gm`/`password`), a Player (`player`/`password`), and a prototype game (`Dragon of Icespire Peak`).
4. Serves HTTP endpoints on port `30011` (or `$PORT`), serving HTMX-infused Tailwind HTML templates dynamically.

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
