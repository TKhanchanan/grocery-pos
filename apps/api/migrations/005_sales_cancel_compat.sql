USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='status') = 0,
  'ALTER TABLE sales ADD COLUMN status ENUM(''COMPLETED'',''CANCELLED'') NOT NULL DEFAULT ''COMPLETED'' AFTER change_amount',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='cancelled_by') = 0,
  'ALTER TABLE sales ADD COLUMN cancelled_by BIGINT UNSIGNED NULL AFTER status',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='cancelled_at') = 0,
  'ALTER TABLE sales ADD COLUMN cancelled_at DATETIME NULL AFTER cancelled_by',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='sales' AND column_name='cancel_reason') = 0,
  'ALTER TABLE sales ADD COLUMN cancel_reason VARCHAR(255) NOT NULL DEFAULT '''' AFTER cancelled_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('005_sales_cancel_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
