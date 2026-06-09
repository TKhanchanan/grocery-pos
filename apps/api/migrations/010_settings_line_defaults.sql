USE grocery_pos;
SET NAMES utf8mb4;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='settings' AND column_name='setting_group') = 0,
  'ALTER TABLE settings ADD COLUMN setting_group VARCHAR(80) NOT NULL DEFAULT ''general'' AFTER setting_value',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='settings' AND column_name='updated_at') = 0,
  'ALTER TABLE settings ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER setting_group',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO settings(setting_key, setting_value, setting_group) VALUES
('shop_name', 'Grocery POS Demo', 'shop'),
('shop_phone', '', 'shop'),
('shop_address', '', 'shop'),
('default_location_id', '', 'shop'),
('receipt_footer', 'ขอบคุณที่อุดหนุน', 'receipt'),
('line_enabled', 'false', 'line'),
('line_token', '', 'line'),
('line_target_id', '', 'line')
ON DUPLICATE KEY UPDATE setting_value=setting_value;

INSERT INTO schema_migrations(version) VALUES ('010_settings_line_defaults')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
