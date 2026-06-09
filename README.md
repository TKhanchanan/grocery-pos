# Grocery POS & Inventory System

Full-stack grocery POS and inventory demo system with multi-location stock, sales, reports, imports/exports, purchase orders, alerts, LINE notification settings, Thai/English UI, and light/dark themes.

## Stack

- Backend: Go HTTP API in `apps/api`
- Frontend: Vue 3 + Vite + TypeScript in `apps/web`
- UI: Tailwind CSS, responsive app shell, reusable components
- State: Pinia
- Routing: Vue Router with role guards
- Database: MySQL 8 via Docker Compose

## Demo Accounts

| Role | Username | Password |
| --- | --- | --- |
| Admin | `admin` | `password` |
| Manager | `manager` | `password` |
| Cashier | `cashier` | `password` |

## Project Structure

```txt
apps/
  api/        Go API, migrations, seed data
  web/        Vue frontend
docs/         Roadmap, tasks, demo script
docker-compose.yml
.env.example
```

The current production-ready app lives under `apps/api` and `apps/web`.

## Run Locally

1. Copy environment values:

```bash
cp .env.example .env
```

2. Start MySQL:

```bash
docker compose up -d mysql
```

3. Start the API:

```bash
cd apps/api
go run ./cmd/server
```

4. Start the web app:

```bash
cd apps/web
npm install
npm run dev
```

5. Open `http://localhost:5173`.

## Useful Endpoints

- `GET http://localhost:8080/api/health`
- `GET http://localhost:8080/api/version`
- API base path: `http://localhost:8080/api/v1`

## Build And QA

```bash
cd apps/api
GOCACHE=/private/tmp/grocery-pos-gocache go test ./...
GOCACHE=/private/tmp/grocery-pos-gocache go build ./...

cd ../web
npm run build
```

Latest verification:

- Backend tests/build: passed
- Frontend TypeScript/Vite build: passed
- Responsive QA: reviewed by page checklist in `docs/demo-script.md`

## Presentation Notes

- Use Admin for the full demo flow.
- Seed data includes Thai locations `หน้าร้าน`, `คลังหลัก` and demo products `ไข่เค็ม`, `มาม่า`, `น้ำอัดลม`.
- CSV exports include a UTF-8 BOM so Thai text opens correctly in Excel.
- LINE notification settings are configurable from Settings. LINE failures are logged and do not break core sale/restock/PO flows.
- Dynamic roles and permissions are managed from the Roles page. Existing `ADMIN`, `MANAGER`, and `CASHIER` users are migrated into system roles by `011_dynamic_rbac.sql`.

## Migration Notes

Fresh Docker initialization runs `001_init.sql`, demo seed data, then compatibility migrations including RBAC, profile avatars, and product images.

For an existing database, apply:

```bash
cd apps/api
go run ./cmd/migrate
```

The migrations keep the legacy `users.role` column for compatibility while adding dynamic RBAC tables, profile avatar fields, and product image fields.
