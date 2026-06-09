USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='note') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN note VARCHAR(255) NOT NULL DEFAULT '''' AFTER total_cost',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='created_by') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN created_by BIGINT UNSIGNED NULL AFTER note',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='received_by') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN received_by BIGINT UNSIGNED NULL AFTER created_by',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='cancelled_by') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN cancelled_by BIGINT UNSIGNED NULL AFTER received_by',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='cancelled_at') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN cancelled_at DATETIME NULL AFTER received_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_orders' AND column_name='updated_at') = 0,
  'ALTER TABLE purchase_orders ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER cancelled_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_order_items' AND column_name='received_quantity') = 0,
  'ALTER TABLE purchase_order_items ADD COLUMN received_quantity INT NOT NULL DEFAULT 0 AFTER quantity',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='purchase_order_items' AND column_name='created_at') = 0,
  'ALTER TABLE purchase_order_items ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER line_cost',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('009_purchase_orders_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
