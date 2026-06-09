USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='stock_movements' AND column_name='before_stock') = 0,
  'ALTER TABLE stock_movements ADD COLUMN before_stock INT NOT NULL DEFAULT 0 AFTER quantity_change',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='stock_movements' AND column_name='after_stock') = 0,
  'ALTER TABLE stock_movements ADD COLUMN after_stock INT NOT NULL DEFAULT 0 AFTER before_stock',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='stock_movements' AND column_name='quantity_after') = 0,
  'ALTER TABLE stock_movements ADD COLUMN quantity_after INT NULL AFTER after_stock',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('003_stock_movements_before_after')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
