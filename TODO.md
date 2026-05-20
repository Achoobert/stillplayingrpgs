# TODO: Finish StartPlaying Clone (Go + HTMX + Docker)

## Phase 1: Go Backend Rewrite & Auth
- [ ] Initialize Go module
- [ ] Implement SQLite database layer with `sqlx` or `gorm`
- [ ] Implement User and Game models
- [ ] Implement Secure Auth (JWT or Session-based with hashes)
- [ ] Port FastAPI endpoints to Go (Echo or Gin)
    - [ ] Public: Landing Page, GM Profile, Login, Register
    - [ ] GM: Dashboard, Create Game, Edit Game
    - [ ] Player: Dashboard, Join Game

## Phase 2: Frontend Refinement
- [ ] Connect Go templates to HTMX
- [ ] Implement toast/error notifications for auth and actions
- [ ] Standardize Tailwind CSS build process

## Phase 3: Containerization & Deployment
- [ ] Create multi-stage `Dockerfile` (Go builder + Alpine runner)
- [ ] Create `docker-compose.yml` for easy RPi deployment
- [ ] Persistent Volume setup for SQLite db

## Phase 4: Cleanup
- [ ] Remove legacy Python files (`src/`, `venv/`, `requirements.txt`)
- [ ] Remove legacy TypeScript files (`package.json`, etc. if not using Hono)
- [ ] Clean up redundant `startplaying-clone/` directory
