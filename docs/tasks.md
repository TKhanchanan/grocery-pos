# Tasks

## Prompt 1: Project Foundation

- [x] Create monorepo structure with `apps/api` and `apps/web`.
- [x] Create Go backend HTTP server.
- [x] Add MySQL connection helper.
- [x] Add config loader.
- [x] Add health endpoint.
- [x] Add version endpoint.
- [x] Add standard JSON response format.
- [x] Add standard error format.
- [x] Add CORS middleware.
- [x] Add request logger middleware.
- [x] Add panic recovery middleware.
- [x] Add transaction helper.
- [x] Add migration folder.
- [x] Add seed folder.
- [x] Add README.
- [x] Add `.env.example`.
- [x] Add `docs/roadmap.md`.
- [x] Add `docs/tasks.md`.
- [x] Add `docs/demo-script.md`.
- [x] Set up Vue 3, Vite, TypeScript, Tailwind CSS, Pinia, and Vue Router.
- [x] Add API client.
- [x] Add Auth layout.
- [x] Add App layout.
- [x] Add responsive sidebar and mobile drawer.
- [x] Add topbar.
- [x] Add user menu placeholder.
- [x] Add alert badge placeholder.
- [x] Add placeholder pages for all requested modules.
- [x] Add requested reusable components.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 2: Full MySQL Schema and Seed Data

- [x] Create full MySQL 8 schema migration.
- [x] Use InnoDB and utf8mb4.
- [x] Use `BIGINT UNSIGNED AUTO_INCREMENT` primary keys.
- [x] Use `DECIMAL(10,2)` money columns.
- [x] Use `DATETIME` timestamps.
- [x] Add foreign keys for important relationships.
- [x] Add indexes for search and report columns.
- [x] Create `users`.
- [x] Create `categories`.
- [x] Create `products`.
- [x] Create `locations`.
- [x] Create `product_stocks`.
- [x] Create `stock_movements`.
- [x] Create `stock_transfers`.
- [x] Create `stock_transfer_items`.
- [x] Create `sales`.
- [x] Create `sale_items`.
- [x] Create `alerts`.
- [x] Create `suppliers`.
- [x] Create `purchase_orders`.
- [x] Create `purchase_order_items`.
- [x] Create `notification_logs`.
- [x] Create `import_jobs`.
- [x] Create `import_job_rows`.
- [x] Create `settings`.
- [x] Keep multi-location stock source of truth in `product_stocks`.
- [x] Keep product name non-unique.
- [x] Keep SKU unique.
- [x] Keep barcode optional and unique when provided.
- [x] Seed admin, manager, and cashier users.
- [x] Seed `หน้าร้าน` and `คลังหลัก` locations.
- [x] Seed `ไข่เค็ม`, `มาม่า`, and `น้ำอัดลม` products.
- [x] Seed demo product stock.
- [x] Seed sample categories.
- [x] Seed sample supplier.
- [x] Seed basic settings.
- [x] Run migration against MySQL.
- [x] Run seed against MySQL.
- [x] Verify Thai text is stored and displayed correctly.

## Prompt 3: Auth, Users, Roles, Security

- [x] Add bcrypt password verification and hashing.
- [x] Add JWT login tokens.
- [x] Add `POST /api/v1/auth/login`.
- [x] Add `POST /api/v1/auth/logout`.
- [x] Add `GET /api/v1/auth/me`.
- [x] Add auth middleware.
- [x] Add role middleware.
- [x] Add `GET /api/v1/users`.
- [x] Add `POST /api/v1/users`.
- [x] Add `GET /api/v1/users/{id}`.
- [x] Add `PATCH /api/v1/users/{id}`.
- [x] Add `PATCH /api/v1/users/{id}/status`.
- [x] Return standard 401 and 403 errors.
- [x] Avoid returning `password_hash`.
- [x] Add login page.
- [x] Add auth Pinia store.
- [x] Add route guard.
- [x] Add role-based sidebar menu.
- [x] Add forbidden page.
- [x] Add users page.
- [x] Add create/edit user form.
- [x] Add disable/enable user action.
- [x] Verify admin login.
- [x] Verify manager login.
- [x] Verify cashier login.
- [x] Verify cashier receives backend 403 for users.
- [x] Verify frontend hides unauthorized menus by role.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 4: Categories, Products, Product Stock, Locations

- [x] Add category APIs.
- [x] Add product APIs.
- [x] Add product status API for deactivate/reactivate.
- [x] Add location APIs.
- [x] Add location status API.
- [x] Add product stock list API.
- [x] Add product stock by product API.
- [x] Validate product price is greater than 0.
- [x] Validate unit cost is greater than or equal to 0.
- [x] Keep threshold and reorder point defaulted to 0.
- [x] Keep product name non-unique.
- [x] Reject duplicate SKU.
- [x] Keep barcode optional and unique when provided.
- [x] Allow cashier product view only.
- [x] Allow admin and manager product creation/editing.
- [x] Add category list/form.
- [x] Add product responsive table and mobile cards.
- [x] Add product create/edit form.
- [x] Add stock status badges.
- [x] Add product stock by location panel.
- [x] Add location list/form.
- [x] Add search by name/SKU/barcode.
- [x] Add filters by category/status/stock status.
- [x] Verify manager can create product.
- [x] Verify cashier can view products.
- [x] Verify cashier receives backend 403 when creating product.
- [x] Verify duplicate SKU is rejected.
- [x] Verify location stock is visible.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 5: Restock, Stock Adjustment, Stock Movements

- [x] Add stock movement before/after migration.
- [x] Add `POST /api/v1/products/{id}/restock`.
- [x] Add `POST /api/v1/products/{id}/adjust-stock`.
- [x] Add `GET /api/v1/stock-movements`.
- [x] Validate restock quantity greater than 0.
- [x] Calculate unit cost from total cost divided by quantity.
- [x] Update product latest unit cost on restock.
- [x] Add stock into selected `product_stocks` location.
- [x] Create `RESTOCK` stock movement with before/after stock.
- [x] Recalculate alerts after restock.
- [x] Require adjustment note.
- [x] Prevent adjustment from making stock negative.
- [x] Create `ADJUSTMENT` stock movement.
- [x] Use database transactions for stock-changing operations.
- [x] Add restock page with product selector and location selector.
- [x] Add current stock display.
- [x] Add quantity and total cost inputs.
- [x] Add unit cost preview.
- [x] Add stock after restock preview.
- [x] Add stock adjustment modal.
- [x] Add stock movement history page.
- [x] Add responsive movement table/cards.
- [x] Verify restock `ไข่เค็ม` 100 units total cost 200 into `หน้าร้าน`.
- [x] Verify stock increases correctly.
- [x] Verify unit cost preview/result is 2 บาท.
- [x] Verify movement history shows before/after stock.
- [x] Verify adjustment cannot make stock negative.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 6: Stock Transfer Between Locations

- [x] Add `GET /api/v1/stock-transfers`.
- [x] Add `POST /api/v1/stock-transfers`.
- [x] Add `GET /api/v1/stock-transfers/{id}`.
- [x] Add `POST /api/v1/stock-transfers/{id}/complete`.
- [x] Add `POST /api/v1/stock-transfers/{id}/cancel`.
- [x] Validate source and destination locations are different.
- [x] Validate source stock before transfer completion.
- [x] Block transfers greater than available stock.
- [x] Complete transfer in a transaction.
- [x] Deduct source location stock.
- [x] Add destination location stock.
- [x] Create `TRANSFER_OUT` movement.
- [x] Create `TRANSFER_IN` movement.
- [x] Prevent negative stock.
- [x] Add transfer list.
- [x] Add transfer form.
- [x] Add source location selector.
- [x] Add destination location selector.
- [x] Add product selector.
- [x] Add quantity input.
- [x] Show available source stock.
- [x] Add transfer detail.
- [x] Add responsive transfer UI.
- [x] Verify transfer `มาม่า` from `คลังหลัก` to `หน้าร้าน`.
- [x] Verify source stock decreases.
- [x] Verify destination stock increases.
- [x] Verify movement history is correct.
- [x] Verify insufficient stock is blocked.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 7: POS, Cart, Barcode, Sale Transaction

- [x] Add `GET /api/v1/pos/products`.
- [x] Add `POST /api/v1/sales`.
- [x] Add `GET /api/v1/sales/{id}/receipt`.
- [x] Search POS products by name, SKU, or barcode.
- [x] Filter POS products by selected location.
- [x] Return stock at selected location.
- [x] Validate authenticated cashier/admin role for sale transaction.
- [x] Validate selected location and cart items.
- [x] Lock `product_stocks` rows during sale.
- [x] Reject insufficient stock.
- [x] Calculate total on backend.
- [x] Reject insufficient payment.
- [x] Create sale and sale item snapshots.
- [x] Deduct stock from selected location.
- [x] Create `SALE` stock movements with before/after stock.
- [x] Recalculate alerts after sale.
- [x] Create `SALE_COMPLETED` notification log.
- [x] Keep stock from becoming negative.
- [x] Add POS page with responsive product search/cards and cart/payment panel.
- [x] Add location selector.
- [x] Add product search by name/SKU/barcode.
- [x] Add manual barcode input.
- [x] Add camera scan modal with manual fallback.
- [x] Add Pinia cart store with required cart actions and totals.
- [x] Add CASH and QR payment methods.
- [x] Add sale success modal and receipt link.
- [x] Add receipt detail page that loads generated receipt by id.
- [x] Verify sale of `มาม่า` 3 items totals 18, paid 20, change 2.
- [x] Verify stock decreases by 3 at selected location.
- [x] Verify receipt is generated.
- [x] Verify insufficient payment is blocked.
- [x] Verify insufficient stock is blocked.
- [x] Verify confirm button disables during submit.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 8: Receipt, Sales History, Cancel Sale

- [x] Add `GET /api/v1/sales`.
- [x] Add `GET /api/v1/sales/{id}`.
- [x] Keep `GET /api/v1/sales/{id}/receipt`.
- [x] Add `POST /api/v1/sales/{id}/cancel`.
- [x] Add filters for date range, cashier, location, payment method, status, and receipt number.
- [x] Restrict sale cancel to Admin/Manager.
- [x] Keep cashier unable to cancel sale.
- [x] Only allow cancelling `COMPLETED` sales.
- [x] Keep cancelled sale records instead of deleting them.
- [x] Store cancel reason.
- [x] Restore stock to original sale location.
- [x] Create `CANCEL_SALE` stock movements.
- [x] Keep cancelled sales visible as `CANCELLED`.
- [x] Add printable receipt card.
- [x] Add print button.
- [x] Add back to POS button.
- [x] Add sales history page.
- [x] Add date/cashier/location/payment/status/receipt filters.
- [x] Add responsive desktop table and mobile cards.
- [x] Add cancel sale button for Admin/Manager.
- [x] Add confirm dialog with cancel reason.
- [x] Add cancelled badge/reason display.
- [x] Verify receipt shows snapshot data.
- [x] Verify sales history filters work.
- [x] Verify cancel restores stock.
- [x] Verify cancelled sale remains visible.
- [x] Verify cashier receives 403 when cancelling sale.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 9: Alerts and Reorder Point

- [x] Add `GET /api/v1/alerts`.
- [x] Add `PATCH /api/v1/alerts/{id}/read`.
- [x] Add `PATCH /api/v1/alerts/read-all`.
- [x] Support alert types `LOW_STOCK`, `OUT_OF_STOCK`, and `REORDER_POINT`.
- [x] Create `OUT_OF_STOCK` when stock is 0.
- [x] Create `LOW_STOCK` when threshold is greater than 0 and stock is less than or equal to threshold.
- [x] Skip `LOW_STOCK` when threshold is 0.
- [x] Create `REORDER_POINT` when reorder point is greater than 0 and stock is less than or equal to reorder point.
- [x] Avoid duplicate unread alerts for the same product/location/type when recalculating.
- [x] Keep alert links to product, restock, and purchase order screens.
- [x] Add alert badge to topbar and sidebar.
- [x] Add Alerts page.
- [x] Add unread/type/location filters.
- [x] Add mark as read.
- [x] Add mark all as read.
- [x] Add responsive alert cards.
- [x] Verify `น้ำอัดลม` stock 5 threshold 10 shows `LOW_STOCK`.
- [x] Verify stock 0 shows `OUT_OF_STOCK`.
- [x] Verify stock less than or equal to reorder point shows `REORDER_POINT`.
- [x] Verify alert badge count updates after reading alerts.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 10: Dashboard and Reports

- [x] Add `GET /api/v1/dashboard/summary`.
- [x] Add `GET /api/v1/reports/daily-sales`.
- [x] Add `GET /api/v1/reports/monthly-sales`.
- [x] Add `GET /api/v1/reports/best-selling`.
- [x] Add `GET /api/v1/reports/profit-by-product`.
- [x] Add `GET /api/v1/reports/stock`.
- [x] Add `GET /api/v1/reports/inventory-valuation`.
- [x] Add `GET /api/v1/reports/payment-summary`.
- [x] Add `GET /api/v1/reports/low-stock`.
- [x] Add `GET /api/v1/reports/reorder`.
- [x] Dashboard returns today sales, today receipts, monthly gross profit, top product, stock alert counts, payment summary, recent sales, and stock sections.
- [x] Reports support date, month, and location filters where applicable.
- [x] Revenue/profit reports exclude cancelled sales.
- [x] Profit is calculated as revenue minus cost.
- [x] Product sales/profit reports use sale item snapshots.
- [x] Add dashboard stat cards.
- [x] Add recent sales section.
- [x] Add low stock section.
- [x] Add top product section.
- [x] Add reports page with tabs.
- [x] Add report filter panel.
- [x] Add responsive report tables/cards.
- [x] Add summary cards.
- [x] Add export placeholder buttons.
- [x] Verify dashboard API and UI build.
- [x] Verify best-selling report works.
- [x] Verify profit per product works.
- [x] Verify stock report works.
- [x] Verify report filters work.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 11: Export Reports

- [x] Add `GET /api/v1/exports/inventory-monthly?month=YYYY-MM&format=csv`.
- [x] Add `GET /api/v1/exports/products?format=csv`.
- [x] Add `GET /api/v1/exports/sales?date_from=&date_to=&format=csv`.
- [x] Add `GET /api/v1/exports/profit?month=YYYY-MM&format=csv`.
- [x] Return direct CSV file responses with attachment filenames.
- [x] Add UTF-8 BOM so Thai text opens correctly in Excel.
- [x] Respect sales date filters.
- [x] Respect profit month and optional location filters.
- [x] Exclude cancelled sales from sales/profit exports.
- [x] Add Exports page.
- [x] Add month/date filters.
- [x] Add download loading state.
- [x] Add export error state.
- [x] Add CSV export buttons on Reports page.
- [x] Keep XLSX as disabled placeholder.
- [x] Verify inventory monthly export works.
- [x] Verify product list export works.
- [x] Verify sales export works.
- [x] Verify profit export works.
- [x] Verify CSV contains Excel-friendly BOM.
- [x] Verify Thai text is readable in CSV response.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 12: Product Import from CSV/Excel

- [x] Add `GET /api/v1/imports/products/template`.
- [x] Add `POST /api/v1/imports/products/preview`.
- [x] Add `POST /api/v1/imports/products/confirm`.
- [x] Add `GET /api/v1/imports`.
- [x] Add `GET /api/v1/imports/{id}`.
- [x] Include template fields `sku`, `name`, `barcode`, `category`, `selling_price`, `unit_cost`, `threshold`, `reorder_point`, `location`, and `initial_stock`.
- [x] Validate SKU is required.
- [x] Validate name is required.
- [x] Validate selling price is greater than 0.
- [x] Validate unit cost is greater than or equal to 0.
- [x] Show duplicate SKU as row-level error.
- [x] Keep invalid rows from being imported.
- [x] Store preview before saving products.
- [x] Require confirm before writing products and stock.
- [x] Create product stock and `IMPORT` stock movement when initial stock is provided.
- [x] Store import jobs and import row errors.
- [x] Add Imports page.
- [x] Add download template button.
- [x] Add upload file area.
- [x] Add preview table.
- [x] Add row-level error display.
- [x] Add confirm import button.
- [x] Add import history/detail.
- [x] Add responsive preview/history cards.
- [x] Verify template download works with Thai text.
- [x] Verify valid file previews and imports.
- [x] Verify invalid rows show errors.
- [x] Verify duplicate SKU is clearly warned.
- [x] Verify import does not save until confirm.
- [x] Backend build verified.
- [x] Frontend build verified.

Note: CSV import is implemented. XLSX uploads are rejected with a clear `400` message for this version.

## Prompt 13: Suppliers and Purchase Orders

- [x] Add supplier APIs.
- [x] Add purchase order APIs.
- [x] Support PO statuses `DRAFT`, `SENT`, `RECEIVED`, and `CANCELLED`.
- [x] Allow manual PO creation.
- [x] Allow PO creation from reorder alerts.
- [x] Receive PO in a transaction.
- [x] Increase stock at the target location on receive.
- [x] Create `PO_RECEIVE` stock movements.
- [x] Keep cancelled PO from affecting stock.
- [x] Add supplier list/form.
- [x] Add PO list, form, detail, send, receive, and cancel actions.
- [x] Add responsive PO UI.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 14: LINE Notifications and Settings

- [x] Add `GET /api/v1/settings`.
- [x] Add `PATCH /api/v1/settings`.
- [x] Add `GET /api/v1/settings/line`.
- [x] Add `PATCH /api/v1/settings/line`.
- [x] Add `POST /api/v1/settings/line/test`.
- [x] Add `GET /api/v1/notification-logs`.
- [x] Add shop profile, receipt, and LINE settings.
- [x] Mask LINE token in the UI.
- [x] Keep LINE failures from breaking sale/restock/PO/transfer transactions.
- [x] Store notification log success/fail/skipped states.
- [x] Restrict Settings to Admin.
- [x] Backend build verified.
- [x] Frontend build verified.

## Prompt 15: Premium UI Foundation

- [x] Add Thai/English language switching.
- [x] Add light/dark theme switching.
- [x] Add text-size preference.
- [x] Add premium design tokens.
- [x] Add app icon system.
- [x] Add collapsible desktop sidebar.
- [x] Add global toast host.
- [x] Upgrade buttons, inputs, selects, textareas, cards, badges, modals, drawers, tables, loading, empty, and error states.
- [x] Add missing reusable components.
- [x] Polish Login, Dashboard, POS, Products, and Settings without changing business logic.
- [x] Frontend build verified.

## Prompt 16: Final Responsive Polish, QA, Demo Data, Deploy Prep

- [x] Review Login, Dashboard, POS, Products, Categories, Restock, Stock Movements, Locations, Transfers, Sales History, Receipt, Alerts, Reports, Exports, Imports, Purchase Orders, Suppliers, Users, and Settings.
- [x] Verify mobile overflow risk and table/card responsiveness by code audit.
- [x] Compact topbar controls for small screens.
- [x] Keep POS mobile sticky cart summary.
- [x] Confirm forms use single-column mobile layouts.
- [x] Confirm tables use horizontal scroll or mobile cards.
- [x] Confirm status badges, important numbers, loading, empty, error, toast, and submit loading states.
- [x] Update README.
- [x] Update demo script with full QA demo flow.
- [x] Update `.env.example`.
- [x] Run backend tests.
- [x] Run backend build.
- [x] Run frontend build.

Note: Browser automation was not available in this session, so responsive QA was completed by code review plus production build verification.

## Prompt 17: Dynamic Role & Permission Management CRUD

- [x] Add dynamic RBAC migration.
- [x] Add `roles`, `permissions`, `role_permissions`, `user_roles`, and `audit_logs`.
- [x] Keep legacy `users.role` column for safe migration.
- [x] Seed system roles `ADMIN`, `MANAGER`, and `CASHIER`.
- [x] Seed stable permission codes by module/action.
- [x] Assign all permissions to `ADMIN`.
- [x] Assign operational permissions to `MANAGER`.
- [x] Assign POS/product/sales/alert permissions to `CASHIER`.
- [x] Migrate existing users into `user_roles`.
- [x] Update `/api/v1/auth/me` to return user, roles, and permissions.
- [x] Add permission middleware and apply it to protected backend endpoints.
- [x] Add role CRUD APIs.
- [x] Add permission list/grouped APIs.
- [x] Add role permission assignment API.
- [x] Add user role assignment API.
- [x] Add audit logging for role and permission changes.
- [x] Add frontend permission helpers.
- [x] Update route guards and sidebar menu visibility to use permission codes.
- [x] Add premium Roles & Permissions page.
- [x] Update Users page for dynamic role assignment.
- [x] Hide/disable key product/category/location actions by permission.
- [x] Backend tests/build verified.
- [x] Frontend build verified.

Note: Existing users should keep access after applying `011_dynamic_rbac.sql`. Users with old sessions should refresh or log in again so `/auth/me` reloads the new permission list.
