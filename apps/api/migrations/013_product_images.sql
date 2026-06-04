USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='products' AND column_name='image_url') = 0,
  'ALTER TABLE products ADD COLUMN image_url VARCHAR(255) NULL AFTER active',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='products' AND column_name='image_path') = 0,
  'ALTER TABLE products ADD COLUMN image_path VARCHAR(500) NULL AFTER image_url',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='products' AND column_name='image_updated_at') = 0,
  'ALTER TABLE products ADD COLUMN image_updated_at DATETIME NULL AFTER image_path',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations(version) VALUES ('013_product_images')
ON DUPLICATE KEY UPDATE version=version;
