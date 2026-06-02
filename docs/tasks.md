# Grocery POS & Inventory System Tasks

## Status

- [x] Project foundation: Go API, Vue 3/Vite/TypeScript frontend, Pinia, Vue Router, Tailwind CSS.
- [x] Database schema and seed data: MySQL 8 InnoDB, utf8mb4, demo users, locations, products, supplier, stock, alerts.
- [x] Auth and role guard: bcrypt password hashes, JWT auth, ADMIN/MANAGER/CASHIER route permissions.
- [x] UI design system and responsive layout: reusable controls, panels, tables, loading/error/empty state component.
- [x] Products and categories: product CRUD V1 with SKU, optional unique barcode, threshold, reorder point.
- [x] Locations and product stocks: multi-location stock using `product_stocks` as source of truth.
- [x] Restock and stock movements: transaction-backed stock increase and movement records.
- [x] Stock adjustment: positive/negative adjustments with non-negative stock enforcement.
- [x] Stock transfer: source/destination movements in one transaction.
- [x] POS cart and sale transaction: sale snapshots, insufficient stock/payment checks, stock row locks, receipt.
- [x] Barcode: manual barcode search plus camera-scan fallback path via decoded input.
- [x] Receipt: receipt detail returned by API and shown after checkout.
- [x] Sales history and cancel sale: cancelled sales retained, stock restored, movement created.
- [x] Alerts: low stock, out-of-stock, and reorder point alerts regenerated after stock changes.
- [x] Dashboard: daily revenue, profit, sale count, items sold, inventory value, alert counts.
- [x] Reports: daily/monthly sales, best-selling, profit per product, stock, valuation, payment summary.
- [x] Export: CSV export endpoints and frontend download.
- [x] Import: CSV product import endpoint and template columns documented in UI.
- [x] Suppliers and purchase orders: supplier CRUD V1, PO creation, PO receive into stock.
- [x] LINE notification: backend hook sends broadcast when `LINE_CHANNEL_ACCESS_TOKEN` is configured.
- [x] Settings: admin settings screen and API.
- [x] Final responsive polish: mobile-friendly navigation, forms, cards, and tables.
- [x] Build/test/deploy docs: README and Docker Compose for MySQL.

## Verification

- [x] Backend build: `GOCACHE=/private/tmp/grocery-pos-gocache GOMODCACHE=/Users/thanyanan/Documents/GitHub/Grocery-POS/backend/.gomodcache go build ./...`
- [x] Frontend build: `npm run build`

## Demo Flow

1. Start MySQL with `docker compose up -d mysql`.
2. Start API from `backend` with `go run ./cmd/api`.
3. Start frontend from `frontend` with `npm run dev`.
4. Login with `admin` / `password`.
5. Open Dashboard, Products, Inventory, POS, Sales, Reports, Suppliers & PO, and Settings.
6. Restock `ไข่เค็ม` 100 units into `หน้าร้าน`, sell 3 units in POS, confirm receipt, verify stock decreases, cancel the sale, and verify stock returns.
7. Export inventory or valuation CSV, create a purchase order from a low/reorder product, receive it, then transfer stock between `คลังหลัก` and `หน้าร้าน`.
