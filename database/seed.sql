USE grocery_pos;
SET NAMES utf8mb4;

INSERT INTO users(username, password_hash, full_name, role, active) VALUES
('admin', '$2a$10$gMHs7AdYx3N7S15WBn9MwONvv3seRxHHs8CADBEx0Kge1lq5uTDXO', 'Admin', 'ADMIN', TRUE),
('manager', '$2a$10$gMHs7AdYx3N7S15WBn9MwONvv3seRxHHs8CADBEx0Kge1lq5uTDXO', 'Manager', 'MANAGER', TRUE),
('cashier', '$2a$10$gMHs7AdYx3N7S15WBn9MwONvv3seRxHHs8CADBEx0Kge1lq5uTDXO', 'Cashier', 'CASHIER', TRUE)
ON DUPLICATE KEY UPDATE password_hash=VALUES(password_hash), full_name=VALUES(full_name), role=VALUES(role), active=VALUES(active);

INSERT INTO categories(name) VALUES ('อาหาร'), ('เครื่องดื่ม'), ('ของใช้ประจำวัน')
ON DUPLICATE KEY UPDATE name=VALUES(name);

INSERT INTO locations(name, active) VALUES ('หน้าร้าน', TRUE), ('คลังหลัก', TRUE)
ON DUPLICATE KEY UPDATE active=VALUES(active);

INSERT INTO products(category_id, sku, barcode, name, unit, price, cost, threshold, reorder_point, active) VALUES
((SELECT id FROM categories WHERE name='อาหาร'), 'EGG-SALTED', '885000000001', 'ไข่เค็ม', 'ฟอง', 5.00, 2.00, 10, 20, TRUE),
((SELECT id FROM categories WHERE name='อาหาร'), 'MAMA-001', '885000000002', 'มาม่า', 'ซอง', 6.00, 4.00, 20, 30, TRUE),
((SELECT id FROM categories WHERE name='เครื่องดื่ม'), 'SODA-001', '885000000003', 'น้ำอัดลม', 'ขวด', 15.00, 10.00, 10, 0, TRUE)
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), cost=VALUES(cost), threshold=VALUES(threshold), reorder_point=VALUES(reorder_point), active=VALUES(active);

INSERT INTO product_stocks(product_id, location_id, quantity) VALUES
((SELECT id FROM products WHERE sku='EGG-SALTED'), (SELECT id FROM locations WHERE name='หน้าร้าน'), 0),
((SELECT id FROM products WHERE sku='EGG-SALTED'), (SELECT id FROM locations WHERE name='คลังหลัก'), 0),
((SELECT id FROM products WHERE sku='MAMA-001'), (SELECT id FROM locations WHERE name='หน้าร้าน'), 100),
((SELECT id FROM products WHERE sku='MAMA-001'), (SELECT id FROM locations WHERE name='คลังหลัก'), 0),
((SELECT id FROM products WHERE sku='SODA-001'), (SELECT id FROM locations WHERE name='หน้าร้าน'), 5),
((SELECT id FROM products WHERE sku='SODA-001'), (SELECT id FROM locations WHERE name='คลังหลัก'), 0)
ON DUPLICATE KEY UPDATE quantity=VALUES(quantity);

DELETE FROM stock_movements WHERE reference_type='SEED';

INSERT INTO stock_movements(product_id, location_id, reference_type, quantity_change, unit_cost, note, created_by)
SELECT p.id, l.id, 'SEED', ps.quantity, p.cost, 'initial demo stock', (SELECT id FROM users WHERE username='admin')
FROM product_stocks ps
JOIN products p ON p.id=ps.product_id
JOIN locations l ON l.id=ps.location_id
WHERE ps.quantity > 0;

INSERT INTO suppliers(name, phone, email, address) VALUES
('Sample Grocery Supplier', '020000000', 'supplier@example.com', 'Bangkok')
ON DUPLICATE KEY UPDATE phone=VALUES(phone), email=VALUES(email), address=VALUES(address);

INSERT INTO settings(setting_key, setting_value) VALUES
('shop_name', 'Grocery POS Demo'),
('currency', 'THB'),
('line_notifications_enabled', 'false')
ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value);

DELETE FROM alerts WHERE resolved_at IS NULL;

INSERT INTO alerts(product_id, location_id, type, message)
SELECT p.id, l.id, 'LOW_STOCK', CONCAT(p.name, ' low stock at ', l.name, ': ', ps.quantity)
FROM products p
JOIN product_stocks ps ON ps.product_id=p.id
JOIN locations l ON l.id=ps.location_id
WHERE p.threshold > 0 AND ps.quantity <= p.threshold;

INSERT INTO alerts(product_id, location_id, type, message)
SELECT p.id, l.id, 'OUT_OF_STOCK', CONCAT(p.name, ' out of stock at ', l.name)
FROM products p
JOIN product_stocks ps ON ps.product_id=p.id
JOIN locations l ON l.id=ps.location_id
WHERE ps.quantity = 0;

INSERT INTO alerts(product_id, location_id, type, message)
SELECT p.id, l.id, 'REORDER_POINT', CONCAT(p.name, ' reached reorder point at ', l.name, ': ', ps.quantity)
FROM products p
JOIN product_stocks ps ON ps.product_id=p.id
JOIN locations l ON l.id=ps.location_id
WHERE p.reorder_point > 0 AND ps.quantity <= p.reorder_point;
