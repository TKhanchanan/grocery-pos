# Demo Script

## Setup

1. Start MySQL with `docker compose up -d mysql`.
2. Start the API:

```bash
cd apps/api
go run ./cmd/server
```

3. Confirm the API responds:

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/version
```

4. Start the frontend:

```bash
cd apps/web
npm install
npm run dev
```

5. Open `http://localhost:5173`.

## Login

1. Login as Admin: `admin` / `password`.
2. Switch language between Thai and English.
3. Switch light/dark theme.
4. Adjust text size from the topbar or Settings.

## Main Demo Flow

1. Open Dashboard and confirm sales, profit, alert, and stock summary cards load.
2. Open Products.
3. Create or edit a product. Confirm duplicate SKU errors are clear.
4. Open Locations and confirm `หน้าร้าน` and `คลังหลัก` exist.
5. Open Restock.
6. Restock `ไข่เค็ม` 100 units into `หน้าร้าน` with total cost 200 บาท.
7. Confirm unit cost preview is 2 บาท and movement history shows before/after stock.
8. Open POS.
9. Select `หน้าร้าน`.
10. Search or barcode-add `ไข่เค็ม`.
11. Sell 3 units, enter enough received cash, and confirm sale.
12. Open the receipt from the success modal.
13. Print-preview the receipt card.
14. Open Alerts and confirm low/out/reorder stock signals are readable.
15. Open Sales History.
16. Filter by receipt or status.
17. Cancel the completed sale with a reason.
18. Confirm the sale remains visible as `CANCELLED` and stock is restored.
19. Open Reports and review best-selling, profit, stock, low-stock, and reorder reports.
20. Open Exports and export inventory, products, sales, and profit CSV files.
21. Open Imports, download the product template, upload a CSV preview, and confirm only after row validation.
22. Open Purchase Orders.
23. Create a PO manually or from a reorder alert.
24. Send and receive the PO.
25. Confirm stock increases and `PO_RECEIVE` movement is created.
26. Open Transfers.
27. Transfer stock from `คลังหลัก` to `หน้าร้าน`.
28. Confirm source stock decreases, destination stock increases, and movements are correct.
29. Open Settings.
30. Configure LINE settings and use Test Send if valid credentials are available.
31. Confirm notification logs show sent/failed/skipped states.

## Responsive QA Checklist

- Login fits on mobile without horizontal overflow.
- Dashboard stat cards wrap cleanly.
- POS works on desktop, tablet, and mobile.
- Mobile POS keeps a sticky cart summary.
- Product, sales, movement, and report tables use scroll or mobile cards.
- Forms become single-column on mobile.
- Buttons are at least touch-friendly height.
- Important numbers use larger type.
- Status badges are visually distinct.
- Thai product/location labels are readable.
- Loading, empty, and error states are visible.
- Toast messages are clear and dismissible.
- Confirm/submit buttons disable or show loading while submitting.
- Resize from mobile to desktop and confirm sidebar/drawer behavior.

## Verified Commands

```bash
cd apps/api
GOCACHE=/private/tmp/grocery-pos-gocache go test ./...
GOCACHE=/private/tmp/grocery-pos-gocache go build ./...

cd ../web
npm run build
```

Current result: all passed.
