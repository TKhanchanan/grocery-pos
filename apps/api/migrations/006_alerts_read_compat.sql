USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='alerts' AND column_name='read_by') = 0,
  'ALTER TABLE alerts ADD COLUMN read_by BIGINT UNSIGNED NULL AFTER message',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='alerts' AND column_name='read_at') = 0,
  'ALTER TABLE alerts ADD COLUMN read_at DATETIME NULL AFTER read_by',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('006_alerts_read_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
