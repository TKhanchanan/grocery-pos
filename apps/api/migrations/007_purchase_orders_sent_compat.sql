USE grocery_pos;
SET NAMES utf8mb4;

UPDATE purchase_orders SET status='SENT' WHERE status='OPEN';

ALTER TABLE purchase_orders
  MODIFY COLUMN status ENUM('DRAFT','SENT','RECEIVED','CANCELLED') NOT NULL DEFAULT 'DRAFT';

INSERT INTO schema_migrations(version) VALUES ('007_purchase_orders_sent_compat')
ON DUPLICATE KEY UPDATE applied_at=applied_at;
