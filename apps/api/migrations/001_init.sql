CREATE DATABASE IF NOT EXISTS grocery_pos
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE grocery_pos;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) PRIMARY KEY,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(80) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(160) NOT NULL,
  role ENUM('ADMIN','MANAGER','CASHIER') NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_users_role (role),
  INDEX idx_users_active (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(160) NOT NULL UNIQUE,
  description VARCHAR(255) NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_categories_active (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  category_id BIGINT UNSIGNED NULL,
  sku VARCHAR(80) NOT NULL UNIQUE,
  barcode VARCHAR(120) NULL UNIQUE,
  name VARCHAR(180) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  unit VARCHAR(40) NOT NULL DEFAULT 'ชิ้น',
  price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  cost DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  threshold INT NOT NULL DEFAULT 0,
  reorder_point INT NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  image_url VARCHAR(255) NULL,
  image_path VARCHAR(500) NULL,
  image_updated_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id),
  INDEX idx_products_name (name),
  INDEX idx_products_category (category_id),
  INDEX idx_products_active (active),
  INDEX idx_products_search (name, sku, barcode)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS locations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(160) NOT NULL UNIQUE,
  description VARCHAR(255) NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_locations_active (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS product_stocks (
  product_id BIGINT UNSIGNED NOT NULL,
  location_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (product_id, location_id),
  CONSTRAINT fk_product_stocks_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT fk_product_stocks_location FOREIGN KEY (location_id) REFERENCES locations(id),
  CONSTRAINT chk_product_stocks_non_negative CHECK (quantity >= 0),
  INDEX idx_product_stocks_location (location_id),
  INDEX idx_product_stocks_quantity (quantity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_movements (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  location_id BIGINT UNSIGNED NOT NULL,
  reference_type VARCHAR(40) NOT NULL,
  reference_id BIGINT UNSIGNED NULL,
  quantity_change INT NOT NULL,
  before_stock INT NOT NULL DEFAULT 0,
  after_stock INT NOT NULL DEFAULT 0,
  quantity_after INT NULL,
  unit_cost DECIMAL(10,2) NULL,
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_stock_movements_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT fk_stock_movements_location FOREIGN KEY (location_id) REFERENCES locations(id),
  CONSTRAINT fk_stock_movements_user FOREIGN KEY (created_by) REFERENCES users(id),
  INDEX idx_stock_movements_product_location (product_id, location_id),
  INDEX idx_stock_movements_reference (reference_type, reference_id),
  INDEX idx_stock_movements_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_transfers (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  transfer_no VARCHAR(40) NOT NULL UNIQUE,
  from_location_id BIGINT UNSIGNED NOT NULL,
  to_location_id BIGINT UNSIGNED NOT NULL,
  status ENUM('DRAFT','COMPLETED','CANCELLED') NOT NULL DEFAULT 'DRAFT',
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NULL,
  completed_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_stock_transfers_from_location FOREIGN KEY (from_location_id) REFERENCES locations(id),
  CONSTRAINT fk_stock_transfers_to_location FOREIGN KEY (to_location_id) REFERENCES locations(id),
  CONSTRAINT fk_stock_transfers_user FOREIGN KEY (created_by) REFERENCES users(id),
  INDEX idx_stock_transfers_status (status),
  INDEX idx_stock_transfers_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_transfer_items (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  transfer_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_stock_transfer_items_transfer FOREIGN KEY (transfer_id) REFERENCES stock_transfers(id),
  CONSTRAINT fk_stock_transfer_items_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT chk_stock_transfer_items_quantity CHECK (quantity > 0),
  INDEX idx_stock_transfer_items_transfer (transfer_id),
  INDEX idx_stock_transfer_items_product (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sales (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  receipt_no VARCHAR(40) NOT NULL UNIQUE,
  location_id BIGINT UNSIGNED NOT NULL,
  cashier_id BIGINT UNSIGNED NOT NULL,
  subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  total_amount DECIMAL(10,2) NOT NULL,
  total_cost DECIMAL(10,2) NOT NULL,
  profit DECIMAL(10,2) NOT NULL,
  payment_method VARCHAR(40) NOT NULL,
  paid_amount DECIMAL(10,2) NOT NULL,
  change_amount DECIMAL(10,2) NOT NULL,
  status ENUM('COMPLETED','CANCELLED') NOT NULL DEFAULT 'COMPLETED',
  cancelled_by BIGINT UNSIGNED NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_sales_location FOREIGN KEY (location_id) REFERENCES locations(id),
  CONSTRAINT fk_sales_cashier FOREIGN KEY (cashier_id) REFERENCES users(id),
  CONSTRAINT fk_sales_cancelled_by FOREIGN KEY (cancelled_by) REFERENCES users(id),
  INDEX idx_sales_created_at (created_at),
  INDEX idx_sales_status_created_at (status, created_at),
  INDEX idx_sales_payment_method (payment_method),
  INDEX idx_sales_location (location_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sale_items (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  sale_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  product_name_snapshot VARCHAR(180) NOT NULL,
  sku_snapshot VARCHAR(80) NOT NULL,
  barcode_snapshot VARCHAR(120) NULL,
  price_snapshot DECIMAL(10,2) NOT NULL,
  cost_snapshot DECIMAL(10,2) NOT NULL,
  quantity INT NOT NULL,
  line_total DECIMAL(10,2) NOT NULL,
  line_cost DECIMAL(10,2) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_sale_items_sale FOREIGN KEY (sale_id) REFERENCES sales(id),
  CONSTRAINT fk_sale_items_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT chk_sale_items_quantity CHECK (quantity > 0),
  INDEX idx_sale_items_sale (sale_id),
  INDEX idx_sale_items_product (product_id),
  INDEX idx_sale_items_report_product (product_id, quantity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS alerts (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  location_id BIGINT UNSIGNED NOT NULL,
  type ENUM('LOW_STOCK','OUT_OF_STOCK','REORDER_POINT') NOT NULL,
  message VARCHAR(255) NOT NULL,
  read_by BIGINT UNSIGNED NULL,
  read_at DATETIME NULL,
  resolved_by BIGINT UNSIGNED NULL,
  resolved_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_alerts_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT fk_alerts_location FOREIGN KEY (location_id) REFERENCES locations(id),
  CONSTRAINT fk_alerts_read_by FOREIGN KEY (read_by) REFERENCES users(id),
  CONSTRAINT fk_alerts_resolved_by FOREIGN KEY (resolved_by) REFERENCES users(id),
  INDEX idx_alerts_unread (read_at, resolved_at),
  INDEX idx_alerts_open (resolved_at, type),
  INDEX idx_alerts_product_location (product_id, location_id, type),
  INDEX idx_alerts_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS suppliers (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(180) NOT NULL,
  phone VARCHAR(60) NOT NULL DEFAULT '',
  email VARCHAR(120) NOT NULL DEFAULT '',
  address VARCHAR(255) NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_suppliers_name (name),
  INDEX idx_suppliers_active (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS purchase_orders (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  po_number VARCHAR(40) NOT NULL UNIQUE,
  supplier_id BIGINT UNSIGNED NOT NULL,
  location_id BIGINT UNSIGNED NOT NULL,
  status ENUM('DRAFT','SENT','RECEIVED','CANCELLED') NOT NULL DEFAULT 'DRAFT',
  total_cost DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NULL,
  received_by BIGINT UNSIGNED NULL,
  cancelled_by BIGINT UNSIGNED NULL,
  received_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_purchase_orders_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
  CONSTRAINT fk_purchase_orders_location FOREIGN KEY (location_id) REFERENCES locations(id),
  CONSTRAINT fk_purchase_orders_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_purchase_orders_received_by FOREIGN KEY (received_by) REFERENCES users(id),
  CONSTRAINT fk_purchase_orders_cancelled_by FOREIGN KEY (cancelled_by) REFERENCES users(id),
  INDEX idx_purchase_orders_status (status),
  INDEX idx_purchase_orders_created_at (created_at),
  INDEX idx_purchase_orders_supplier (supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS purchase_order_items (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  po_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL,
  received_quantity INT NOT NULL DEFAULT 0,
  unit_cost DECIMAL(10,2) NOT NULL,
  line_cost DECIMAL(10,2) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_purchase_order_items_po FOREIGN KEY (po_id) REFERENCES purchase_orders(id),
  CONSTRAINT fk_purchase_order_items_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT chk_purchase_order_items_quantity CHECK (quantity > 0),
  CONSTRAINT chk_purchase_order_items_received CHECK (received_quantity >= 0),
  INDEX idx_purchase_order_items_po (po_id),
  INDEX idx_purchase_order_items_product (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS notification_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  channel VARCHAR(40) NOT NULL,
  recipient VARCHAR(160) NOT NULL DEFAULT '',
  event_type VARCHAR(80) NOT NULL,
  payload JSON NULL,
  status ENUM('PENDING','SENT','FAILED','SKIPPED') NOT NULL DEFAULT 'PENDING',
  error_message VARCHAR(255) NOT NULL DEFAULT '',
  sent_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_notification_logs_channel_status (channel, status),
  INDEX idx_notification_logs_event_type (event_type),
  INDEX idx_notification_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS import_jobs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  job_type VARCHAR(60) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  status ENUM('PENDING','PROCESSING','COMPLETED','FAILED') NOT NULL DEFAULT 'PENDING',
  total_rows INT NOT NULL DEFAULT 0,
  success_rows INT NOT NULL DEFAULT 0,
  failed_rows INT NOT NULL DEFAULT 0,
  created_by BIGINT UNSIGNED NULL,
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_import_jobs_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  INDEX idx_import_jobs_status (status),
  INDEX idx_import_jobs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS import_job_rows (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  import_job_id BIGINT UNSIGNED NOT NULL,
  row_index INT NOT NULL,
  raw_data JSON NULL,
  status ENUM('PENDING','IMPORTED','FAILED') NOT NULL DEFAULT 'PENDING',
  error_message VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_import_job_rows_job FOREIGN KEY (import_job_id) REFERENCES import_jobs(id),
  INDEX idx_import_job_rows_job (import_job_id),
  INDEX idx_import_job_rows_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS settings (
  setting_key VARCHAR(100) PRIMARY KEY,
  setting_value TEXT NOT NULL,
  setting_group VARCHAR(80) NOT NULL DEFAULT 'general',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_settings_group (setting_group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO schema_migrations(version) VALUES ('001_init')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
