USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='subtotal') = 0,
  'ALTER TABLE sales ADD COLUMN subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER cashier_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='discount_amount') = 0,
  'ALTER TABLE sales ADD COLUMN discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER subtotal',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='total_amount') = 0,
  'ALTER TABLE sales ADD COLUMN total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER discount_amount',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='total_cost') = 0,
  'ALTER TABLE sales ADD COLUMN total_cost DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER total_amount',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='profit') = 0,
  'ALTER TABLE sales ADD COLUMN profit DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER total_cost',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='paid_amount') = 0,
  'ALTER TABLE sales ADD COLUMN paid_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER payment_method',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='change_amount') = 0,
  'ALTER TABLE sales ADD COLUMN change_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER paid_amount',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='product_name_snapshot') = 0,
  'ALTER TABLE sale_items ADD COLUMN product_name_snapshot VARCHAR(180) NOT NULL DEFAULT '''' AFTER product_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='sku_snapshot') = 0,
  'ALTER TABLE sale_items ADD COLUMN sku_snapshot VARCHAR(80) NOT NULL DEFAULT '''' AFTER product_name_snapshot',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='barcode_snapshot') = 0,
  'ALTER TABLE sale_items ADD COLUMN barcode_snapshot VARCHAR(120) NULL AFTER sku_snapshot',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='price_snapshot') = 0,
  'ALTER TABLE sale_items ADD COLUMN price_snapshot DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER barcode_snapshot',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='cost_snapshot') = 0,
  'ALTER TABLE sale_items ADD COLUMN cost_snapshot DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER price_snapshot',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='line_total') = 0,
  'ALTER TABLE sale_items ADD COLUMN line_total DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER quantity',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sale_items' AND column_name='line_cost') = 0,
  'ALTER TABLE sale_items ADD COLUMN line_cost DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER line_total',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('004_sales_pos_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
