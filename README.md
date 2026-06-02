# Grocery POS & Inventory System

Full-scope V1 for a small grocery shop with Go, MySQL 8, Vue 3, Vite, TypeScript, Pinia, Vue Router, and Tailwind CSS.

## Demo Login

All demo users use password `password`.

- `admin` / `ADMIN`
- `manager` / `MANAGER`
- `cashier` / `CASHIER`

## Run Locally

1. Start MySQL:

```bash
docker compose up -d mysql
```

2. Start backend:

```bash
cd backend
go mod download
go run ./cmd/api
```

3. Start frontend:

```bash
cd frontend
npm install
npm run dev
```

The frontend expects the API at `http://localhost:8080/api`. Set `VITE_API_URL` to override it.

## Demo Path

Login as `admin`, open Dashboard, edit a product, restock `ไข่เค็ม` into `หน้าร้าน`, open POS, sell 3 eggs, print/view receipt, inspect stock movement and alerts, cancel the sale, export reports, import product CSV, create and receive a purchase order, transfer stock, and resize the UI.
