USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='phone') = 0,
  'ALTER TABLE suppliers ADD COLUMN phone VARCHAR(60) NOT NULL DEFAULT '''' AFTER name',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='email') = 0,
  'ALTER TABLE suppliers ADD COLUMN email VARCHAR(120) NOT NULL DEFAULT '''' AFTER phone',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='address') = 0,
  'ALTER TABLE suppliers ADD COLUMN address VARCHAR(255) NOT NULL DEFAULT '''' AFTER email',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='active') = 0,
  'ALTER TABLE suppliers ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE AFTER address',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='created_at') = 0,
  'ALTER TABLE suppliers ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER active',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='suppliers' AND column_name='updated_at') = 0,
  'ALTER TABLE suppliers ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('008_purchasing_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
