USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='avatar_url') = 0,
  'ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255) NULL AFTER active',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='avatar_path') = 0,
  'ALTER TABLE users ADD COLUMN avatar_path VARCHAR(500) NULL AFTER avatar_url',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='avatar_updated_at') = 0,
  'ALTER TABLE users ADD COLUMN avatar_updated_at DATETIME NULL AFTER avatar_path',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('012_profile_avatar')
ON DUPLICATE KEY UPDATE version=version;
