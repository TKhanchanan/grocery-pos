USE grocery_pos;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(80) NOT NULL UNIQUE,
  name VARCHAR(160) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  is_system BOOLEAN NOT NULL DEFAULT FALSE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_roles_active (is_active),
  INDEX idx_roles_system (is_system)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(120) NOT NULL UNIQUE,
  module VARCHAR(80) NOT NULL,
  action VARCHAR(80) NOT NULL,
  name VARCHAR(160) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_permissions_module_sort (module, sort_order),
  INDEX idx_permissions_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  role_id BIGINT UNSIGNED NOT NULL,
  permission_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id),
  CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id),
  UNIQUE KEY uniq_role_permissions_pair (role_id, permission_id),
  INDEX idx_role_permissions_permission (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id),
  UNIQUE KEY uniq_user_roles_pair (user_id, role_id),
  INDEX idx_user_roles_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  actor_user_id BIGINT UNSIGNED NULL,
  action VARCHAR(80) NOT NULL,
  entity_type VARCHAR(80) NOT NULL,
  entity_id BIGINT UNSIGNED NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_audit_logs_actor FOREIGN KEY (actor_user_id) REFERENCES users(id),
  INDEX idx_audit_logs_actor (actor_user_id),
  INDEX idx_audit_logs_entity (entity_type, entity_id),
  INDEX idx_audit_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO roles(code, name, description, is_system, is_active) VALUES
('ADMIN', 'Admin', 'Full system access', TRUE, TRUE),
('MANAGER', 'Manager', 'Operational management access', TRUE, TRUE),
('CASHIER', 'Cashier', 'POS and cashier access', TRUE, TRUE)
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), is_system=TRUE, is_active=TRUE;

INSERT INTO permissions(code, module, action, name, description, sort_order) VALUES
('dashboard.view', 'dashboard', 'view', 'View dashboard', '', 10),
('pos.view', 'pos', 'view', 'View POS', '', 20),
('pos.sell', 'pos', 'sell', 'Sell products', '', 21),
('pos.clear_cart', 'pos', 'clear_cart', 'Clear POS cart', '', 22),
('pos.apply_discount', 'pos', 'apply_discount', 'Apply discount', '', 23),
('products.view', 'products', 'view', 'View products', '', 30),
('products.create', 'products', 'create', 'Create products', '', 31),
('products.update', 'products', 'update', 'Update products', '', 32),
('products.deactivate', 'products', 'deactivate', 'Deactivate products', '', 33),
('products.import', 'products', 'import', 'Import products', '', 34),
('products.export', 'products', 'export', 'Export products', '', 35),
('categories.view', 'categories', 'view', 'View categories', '', 40),
('categories.create', 'categories', 'create', 'Create categories', '', 41),
('categories.update', 'categories', 'update', 'Update categories', '', 42),
('categories.deactivate', 'categories', 'deactivate', 'Deactivate categories', '', 43),
('stock.view', 'stock', 'view', 'View stock', '', 50),
('stock.restock', 'stock', 'restock', 'Restock products', '', 51),
('stock.adjust', 'stock', 'adjust', 'Adjust stock', '', 52),
('stock.movements.view', 'stock', 'movements.view', 'View stock movements', '', 53),
('locations.view', 'locations', 'view', 'View locations', '', 60),
('locations.create', 'locations', 'create', 'Create locations', '', 61),
('locations.update', 'locations', 'update', 'Update locations', '', 62),
('locations.deactivate', 'locations', 'deactivate', 'Deactivate locations', '', 63),
('transfers.view', 'transfers', 'view', 'View transfers', '', 70),
('transfers.create', 'transfers', 'create', 'Create transfers', '', 71),
('transfers.complete', 'transfers', 'complete', 'Complete transfers', '', 72),
('transfers.cancel', 'transfers', 'cancel', 'Cancel transfers', '', 73),
('sales.view', 'sales', 'view', 'View sales', '', 80),
('sales.receipt.view', 'sales', 'receipt.view', 'View receipts', '', 81),
('sales.cancel', 'sales', 'cancel', 'Cancel sales', '', 82),
('alerts.view', 'alerts', 'view', 'View alerts', '', 90),
('alerts.mark_read', 'alerts', 'mark_read', 'Mark alerts read', '', 91),
('alerts.create_po', 'alerts', 'create_po', 'Create PO from alerts', '', 92),
('reports.view', 'reports', 'view', 'View reports', '', 100),
('reports.daily_sales', 'reports', 'daily_sales', 'Daily sales report', '', 101),
('reports.monthly_sales', 'reports', 'monthly_sales', 'Monthly sales report', '', 102),
('reports.best_selling', 'reports', 'best_selling', 'Best-selling report', '', 103),
('reports.profit', 'reports', 'profit', 'Profit report', '', 104),
('reports.stock', 'reports', 'stock', 'Stock report', '', 105),
('reports.inventory_valuation', 'reports', 'inventory_valuation', 'Inventory valuation report', '', 106),
('reports.payment_summary', 'reports', 'payment_summary', 'Payment summary report', '', 107),
('reports.low_stock', 'reports', 'low_stock', 'Low-stock report', '', 108),
('reports.reorder', 'reports', 'reorder', 'Reorder report', '', 109),
('exports.view', 'exports', 'view', 'View exports', '', 110),
('exports.inventory', 'exports', 'inventory', 'Export inventory', '', 111),
('exports.products', 'exports', 'products', 'Export products', '', 112),
('exports.sales', 'exports', 'sales', 'Export sales', '', 113),
('exports.profit', 'exports', 'profit', 'Export profit', '', 114),
('imports.view', 'imports', 'view', 'View imports', '', 120),
('imports.template.download', 'imports', 'template.download', 'Download import template', '', 121),
('imports.products.preview', 'imports', 'products.preview', 'Preview product import', '', 122),
('imports.products.confirm', 'imports', 'products.confirm', 'Confirm product import', '', 123),
('imports.history.view', 'imports', 'history.view', 'View import history', '', 124),
('suppliers.view', 'suppliers', 'view', 'View suppliers', '', 130),
('suppliers.create', 'suppliers', 'create', 'Create suppliers', '', 131),
('suppliers.update', 'suppliers', 'update', 'Update suppliers', '', 132),
('suppliers.deactivate', 'suppliers', 'deactivate', 'Deactivate suppliers', '', 133),
('purchase_orders.view', 'purchase_orders', 'view', 'View purchase orders', '', 140),
('purchase_orders.create', 'purchase_orders', 'create', 'Create purchase orders', '', 141),
('purchase_orders.update', 'purchase_orders', 'update', 'Update purchase orders', '', 142),
('purchase_orders.send', 'purchase_orders', 'send', 'Send purchase orders', '', 143),
('purchase_orders.receive', 'purchase_orders', 'receive', 'Receive purchase orders', '', 144),
('purchase_orders.cancel', 'purchase_orders', 'cancel', 'Cancel purchase orders', '', 145),
('purchase_orders.create_from_alert', 'purchase_orders', 'create_from_alert', 'Create PO from alert', '', 146),
('users.view', 'users', 'view', 'View users', '', 150),
('users.create', 'users', 'create', 'Create users', '', 151),
('users.update', 'users', 'update', 'Update users', '', 152),
('users.deactivate', 'users', 'deactivate', 'Deactivate users', '', 153),
('users.assign_roles', 'users', 'assign_roles', 'Assign user roles', '', 154),
('roles.view', 'roles', 'view', 'View roles', '', 160),
('roles.create', 'roles', 'create', 'Create roles', '', 161),
('roles.update', 'roles', 'update', 'Update roles', '', 162),
('roles.deactivate', 'roles', 'deactivate', 'Deactivate roles', '', 163),
('roles.assign_permissions', 'roles', 'assign_permissions', 'Assign role permissions', '', 164),
('permissions.view', 'permissions', 'view', 'View permissions', '', 165),
('settings.view', 'settings', 'view', 'View settings', '', 170),
('settings.update', 'settings', 'update', 'Update settings', '', 171),
('settings.line.view', 'settings', 'line.view', 'View LINE settings', '', 172),
('settings.line.update', 'settings', 'line.update', 'Update LINE settings', '', 173),
('settings.line.test', 'settings', 'line.test', 'Test LINE settings', '', 174),
('notifications.view', 'notifications', 'view', 'View notifications', '', 180)
ON DUPLICATE KEY UPDATE module=VALUES(module), action=VALUES(action), name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order);

INSERT IGNORE INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p WHERE r.code = 'ADMIN';

INSERT IGNORE INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'MANAGER'
  AND p.code IN (
    'dashboard.view','pos.view','pos.sell',
    'products.view','products.create','products.update','products.deactivate','products.import','products.export',
    'categories.view','categories.create','categories.update','categories.deactivate',
    'stock.view','stock.restock','stock.adjust','stock.movements.view',
    'locations.view','locations.create','locations.update','locations.deactivate',
    'transfers.view','transfers.create','transfers.complete','transfers.cancel',
    'sales.view','sales.receipt.view','sales.cancel',
    'alerts.view','alerts.mark_read','alerts.create_po',
    'reports.view','reports.daily_sales','reports.monthly_sales','reports.best_selling','reports.profit','reports.stock','reports.inventory_valuation','reports.payment_summary','reports.low_stock','reports.reorder',
    'exports.view','exports.inventory','exports.products','exports.sales','exports.profit',
    'imports.view','imports.template.download','imports.products.preview','imports.products.confirm','imports.history.view',
    'suppliers.view','suppliers.create','suppliers.update','suppliers.deactivate',
    'purchase_orders.view','purchase_orders.create','purchase_orders.update','purchase_orders.send','purchase_orders.receive','purchase_orders.cancel','purchase_orders.create_from_alert',
    'notifications.view'
  );

INSERT IGNORE INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'CASHIER'
  AND p.code IN (
    'dashboard.view','pos.view','pos.sell','pos.clear_cart','products.view',
    'sales.view','sales.receipt.view','alerts.view','alerts.mark_read'
  );

INSERT IGNORE INTO user_roles(user_id, role_id)
SELECT u.id, r.id
FROM users u JOIN roles r ON r.code = u.role;

INSERT INTO schema_migrations(version) VALUES ('011_dynamic_rbac')
ON DUPLICATE KEY UPDATE version=version;
