# Grocery POS & Inventory System

Monorepo foundation for a small grocery POS and inventory system.

## Stack

- Backend: Go in `apps/api`
- Frontend: Vue 3 + Vite + TypeScript in `apps/web`
- UI: Tailwind CSS
- State: Pinia
- Routing: Vue Router
- Database: MySQL 8 via Docker Compose

## Structure

```txt
apps/
  api/        Go HTTP API foundation
  web/        Vue frontend foundation
docs/         Roadmap, tasks, demo script
docker-compose.yml
.env.example
```

## Run MySQL

```bash
docker compose up -d mysql
```

## Run Backend

```bash
cd apps/api
go run ./cmd/server
```

Useful endpoints:

- `GET http://localhost:8080/api/health`
- `GET http://localhost:8080/api/version`

## Run Frontend

```bash
cd apps/web
npm install
npm run dev
```

The web app defaults to `http://localhost:5173` and calls the API at `http://localhost:8080/api`.
