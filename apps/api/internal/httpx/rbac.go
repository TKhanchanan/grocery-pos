package httpx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type RoleRecord struct {
	ID              uint64    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsSystem        bool      `json:"is_system"`
	IsActive        bool      `json:"is_active"`
	PermissionCount int       `json:"permission_count"`
	UserCount       int       `json:"user_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PermissionRecord struct {
	ID          uint64    `json:"id"`
	Code        string    `json:"code"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (s *Server) roles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		roles, err := s.roleList(r.Context())
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "ROLES_LOAD_FAILED", "Could not load roles.")
			return
		}
		response.JSON(w, http.StatusOK, roles)
	case http.MethodPost:
		var body RoleInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		role, err := s.createRole(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "ROLE_CREATE_FAILED", err.Error())
			return
		}
		s.audit(r.Context(), "ROLE_CREATED", "role", role.ID, nil, role)
		response.JSON(w, http.StatusCreated, role)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) roleDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid role id.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		role, err := s.roleByID(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found.")
			return
		}
		response.JSON(w, http.StatusOK, role)
	case http.MethodPatch:
		var body RoleInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		before, _ := s.roleByID(r.Context(), id)
		role, err := s.updateRole(r.Context(), id, body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "ROLE_UPDATE_FAILED", err.Error())
			return
		}
		s.audit(r.Context(), "ROLE_UPDATED", "role", role.ID, before, role)
		response.JSON(w, http.StatusOK, role)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) roleStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid role id.")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if !body.IsActive && s.roleIsLastAdminCapable(r.Context(), id) {
		response.ErrorJSON(w, http.StatusBadRequest, "LAST_ADMIN_ROLE", "Cannot deactivate the last admin-capable role.")
		return
	}
	before, _ := s.roleByID(r.Context(), id)
	if _, err := s.db.ExecContext(r.Context(), `UPDATE roles SET is_active=? WHERE id=?`, body.IsActive, id); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "ROLE_STATUS_FAILED", "Could not update role status.")
		return
	}
	role, err := s.roleByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found.")
		return
	}
	s.audit(r.Context(), "ROLE_DEACTIVATED", "role", role.ID, before, role)
	response.JSON(w, http.StatusOK, role)
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid role id.")
		return
	}
	role, err := s.roleByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found.")
		return
	}
	if role.IsSystem || role.UserCount > 0 || s.roleIsLastAdminCapable(r.Context(), id) {
		if _, err := s.db.ExecContext(r.Context(), `UPDATE roles SET is_active=FALSE WHERE id=?`, id); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "ROLE_DELETE_FAILED", "Could not deactivate role.")
			return
		}
		updated, _ := s.roleByID(r.Context(), id)
		s.audit(r.Context(), "ROLE_DEACTIVATED", "role", id, role, updated)
		response.JSON(w, http.StatusOK, updated)
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `DELETE FROM roles WHERE id=?`, id); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "ROLE_DELETE_FAILED", "Could not delete role.")
		return
	}
	s.audit(r.Context(), "ROLE_DELETED", "role", id, role, nil)
	response.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (s *Server) permissions(w http.ResponseWriter, r *http.Request) {
	items, err := s.permissionList(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PERMISSIONS_LOAD_FAILED", "Could not load permissions.")
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (s *Server) groupedPermissions(w http.ResponseWriter, r *http.Request) {
	items, err := s.permissionList(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PERMISSIONS_LOAD_FAILED", "Could not load permissions.")
		return
	}
	grouped := map[string][]PermissionRecord{}
	for _, item := range items {
		grouped[item.Module] = append(grouped[item.Module], item)
	}
	response.JSON(w, http.StatusOK, grouped)
}

func (s *Server) rolePermissions(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid role id.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		codes, err := s.permissionsForRole(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found.")
			return
		}
		response.JSON(w, http.StatusOK, codes)
	case http.MethodPut:
		var body struct {
			PermissionCodes []string `json:"permission_codes"`
		}
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		before, _ := s.permissionsForRole(r.Context(), id)
		codes, err := s.replaceRolePermissions(r.Context(), id, body.PermissionCodes)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "ROLE_PERMISSIONS_FAILED", err.Error())
			return
		}
		s.audit(r.Context(), "ROLE_PERMISSIONS_UPDATED", "role", id, before, codes)
		response.JSON(w, http.StatusOK, codes)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) userRoles(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id.")
		return
	}
	user, err := s.userByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		response.JSON(w, http.StatusOK, s.rolesForUser(r.Context(), user))
	case http.MethodPut:
		var body struct {
			RoleIDs []uint64 `json:"role_ids"`
		}
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		before := s.rolesForUser(r.Context(), user)
		if err := s.replaceUserRoles(r.Context(), id, body.RoleIDs, user.Role); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "USER_ROLES_FAILED", err.Error())
			return
		}
		after := s.rolesForUser(r.Context(), user)
		s.audit(r.Context(), "USER_ROLES_UPDATED", "user", id, before, after)
		response.JSON(w, http.StatusOK, after)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) roleList(ctx context.Context) ([]RoleRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.code, r.name, r.description, r.is_system, r.is_active, r.created_at, r.updated_at,
		       COUNT(DISTINCT rp.permission_id), COUNT(DISTINCT ur.user_id)
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN user_roles ur ON ur.role_id = r.id
		GROUP BY r.id
		ORDER BY r.is_system DESC, r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []RoleRecord{}
	for rows.Next() {
		var role RoleRecord
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.IsSystem, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.PermissionCount, &role.UserCount); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *Server) roleByID(ctx context.Context, id uint64) (RoleRecord, error) {
	var role RoleRecord
	err := s.db.QueryRowContext(ctx, `
		SELECT r.id, r.code, r.name, r.description, r.is_system, r.is_active, r.created_at, r.updated_at,
		       COUNT(DISTINCT rp.permission_id), COUNT(DISTINCT ur.user_id)
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN user_roles ur ON ur.role_id = r.id
		WHERE r.id=?
		GROUP BY r.id`, id).
		Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.IsSystem, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.PermissionCount, &role.UserCount)
	return role, err
}

func (s *Server) createRole(ctx context.Context, body RoleInput) (RoleRecord, error) {
	code := normalizeRoleCode(body.Code)
	if code == "" || strings.TrimSpace(body.Name) == "" {
		return RoleRecord{}, errors.New("role code and name are required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO roles(code, name, description, is_system, is_active) VALUES (?, ?, ?, FALSE, ?)`, code, strings.TrimSpace(body.Name), strings.TrimSpace(body.Description), body.IsActive)
	if err != nil {
		return RoleRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.roleByID(ctx, uint64(id))
}

func (s *Server) updateRole(ctx context.Context, id uint64, body RoleInput) (RoleRecord, error) {
	if strings.TrimSpace(body.Name) == "" {
		return RoleRecord{}, errors.New("role name is required")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE roles SET name=?, description=?, is_active=? WHERE id=?`, strings.TrimSpace(body.Name), strings.TrimSpace(body.Description), body.IsActive, id); err != nil {
		return RoleRecord{}, err
	}
	return s.roleByID(ctx, id)
}

func (s *Server) permissionList(ctx context.Context) ([]PermissionRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, code, module, action, name, description, sort_order, created_at, updated_at FROM permissions ORDER BY module, sort_order, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PermissionRecord{}
	for rows.Next() {
		var item PermissionRecord
		if err := rows.Scan(&item.ID, &item.Code, &item.Module, &item.Action, &item.Name, &item.Description, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Server) permissionsForRole(ctx context.Context, roleID uint64) ([]string, error) {
	if _, err := s.roleByID(ctx, roleID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.code
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id=?
		ORDER BY p.module, p.sort_order, p.code`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (s *Server) replaceRolePermissions(ctx context.Context, roleID uint64, permissionCodes []string) ([]string, error) {
	role, err := s.roleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	codes := normalizePermissionCodes(permissionCodes)
	if role.Code == "ADMIN" && !contains(codes, "roles.assign_permissions") {
		return nil, errors.New("ADMIN role must keep roles.assign_permissions")
	}
	if role.Code == "ADMIN" && !contains(codes, "users.assign_roles") {
		return nil, errors.New("ADMIN role must keep users.assign_roles")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, roleID); err != nil {
		return nil, err
	}
	for _, code := range codes {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO role_permissions(role_id, permission_id)
			SELECT ?, id FROM permissions WHERE code=?`, roleID, code)
		if err != nil {
			return nil, err
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return nil, errors.New("unknown permission code: " + code)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if s.noAdminCapableUsers(ctx) {
		return nil, errors.New("at least one active user must keep admin role management permissions")
	}
	return s.permissionsForRole(ctx, roleID)
}

func (s *Server) rolesForUser(ctx context.Context, user User) []UserRole {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.code, r.name
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id=? AND r.is_active=TRUE
		ORDER BY r.name`, user.ID)
	if err != nil {
		return fallbackUserRoles(user)
	}
	defer rows.Close()
	roles := []UserRole{}
	for rows.Next() {
		var role UserRole
		if err := rows.Scan(&role.ID, &role.Code, &role.Name); err != nil {
			return fallbackUserRoles(user)
		}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return fallbackUserRoles(user)
	}
	return roles
}

func (s *Server) permissionsForUser(ctx context.Context, user User) []string {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT p.code
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id AND r.is_active=TRUE
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id=?
		ORDER BY p.code`, user.ID)
	if err != nil {
		return fallbackPermissions(user.Role)
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return fallbackPermissions(user.Role)
		}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return fallbackPermissions(user.Role)
	}
	return codes
}

func (s *Server) userHasPermission(ctx context.Context, user User, permission string) bool {
	for _, code := range s.permissionsForUser(ctx, user) {
		if code == permission {
			return true
		}
	}
	return false
}

func (s *Server) assignRolesToUser(ctx context.Context, userID uint64, roleIDs []uint64, legacy Role) error {
	err := s.replaceUserRoles(ctx, userID, roleIDs, legacy)
	if err != nil && isMissingRBACTable(err) {
		return nil
	}
	return err
}

func (s *Server) replaceUserRoles(ctx context.Context, userID uint64, roleIDs []uint64, legacy Role) error {
	if len(roleIDs) == 0 {
		roleID, err := s.roleIDByCode(ctx, string(legacy))
		if err != nil {
			return err
		}
		roleIDs = []uint64{roleID}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, userID); err != nil {
		return err
	}
	for _, roleID := range uniqueUint64(roleIDs) {
		result, err := tx.ExecContext(ctx, `INSERT INTO user_roles(user_id, role_id) SELECT ?, id FROM roles WHERE id=? AND is_active=TRUE`, userID, roleID)
		if err != nil {
			return err
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return errors.New("role is inactive or missing")
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if s.noAdminCapableUsers(ctx) {
		return errors.New("at least one active user must keep admin role management permissions")
	}
	return nil
}

func (s *Server) roleIDByCode(ctx context.Context, code string) (uint64, error) {
	var id uint64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM roles WHERE code=? AND is_active=TRUE`, code).Scan(&id)
	return id, err
}

func (s *Server) roleIsLastAdminCapable(ctx context.Context, roleID uint64) bool {
	if !s.roleHasAdminPower(ctx, roleID) {
		return false
	}
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT r.id)
		FROM roles r
		JOIN role_permissions rp1 ON rp1.role_id = r.id
		JOIN permissions p1 ON p1.id = rp1.permission_id AND p1.code='roles.assign_permissions'
		JOIN role_permissions rp2 ON rp2.role_id = r.id
		JOIN permissions p2 ON p2.id = rp2.permission_id AND p2.code='users.assign_roles'
		WHERE r.is_active=TRUE AND r.id<>?`, roleID).Scan(&count)
	return err == nil && count == 0
}

func (s *Server) roleHasAdminPower(ctx context.Context, roleID uint64) bool {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT p.code)
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id=? AND p.code IN ('roles.assign_permissions','users.assign_roles')`, roleID).Scan(&count)
	return err == nil && count >= 2
}

func (s *Server) noAdminCapableUsers(ctx context.Context) bool {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id AND r.is_active=TRUE
		JOIN role_permissions rp1 ON rp1.role_id = r.id
		JOIN permissions p1 ON p1.id = rp1.permission_id AND p1.code='roles.assign_permissions'
		JOIN role_permissions rp2 ON rp2.role_id = r.id
		JOIN permissions p2 ON p2.id = rp2.permission_id AND p2.code='users.assign_roles'
		WHERE u.active=TRUE`).Scan(&count)
	return err == nil && count == 0
}

func (s *Server) audit(ctx context.Context, action, entityType string, entityID uint64, before, after any) {
	actorID := sql.NullInt64{}
	if user, ok := currentUser(ctx); ok {
		actorID.Valid = true
		actorID.Int64 = int64(user.ID)
	}
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	_, _ = s.db.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id, action, entity_type, entity_id, before_json, after_json) VALUES (?, ?, ?, ?, ?, ?)`, actorID, action, entityType, entityID, nullableJSON(beforeJSON, before), nullableJSON(afterJSON, after))
}

func nullableJSON(data []byte, value any) any {
	if value == nil {
		return nil
	}
	return string(data)
}

func normalizeRoleCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, ch := range value {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func normalizePermissionCodes(codes []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func uniqueUint64(values []uint64) []uint64 {
	seen := map[uint64]bool{}
	out := []uint64{}
	for _, value := range values {
		if value == 0 || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isMissingRBACTable(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "doesn't exist") || strings.Contains(message, "no such table")
}

func fallbackUserRoles(user User) []UserRole {
	return []UserRole{{ID: 0, Code: string(user.Role), Name: strings.Title(strings.ToLower(string(user.Role)))}}
}

func fallbackPermissions(role Role) []string {
	admin := []string{
		"dashboard.view", "pos.view", "pos.sell", "pos.clear_cart", "pos.apply_discount",
		"products.view", "products.create", "products.update", "products.deactivate", "products.import", "products.export",
		"categories.view", "categories.create", "categories.update", "categories.deactivate",
		"stock.view", "stock.restock", "stock.adjust", "stock.movements.view",
		"locations.view", "locations.create", "locations.update", "locations.deactivate",
		"transfers.view", "transfers.create", "transfers.complete", "transfers.cancel",
		"sales.view", "sales.receipt.view", "sales.cancel",
		"alerts.view", "alerts.mark_read", "alerts.create_po",
		"reports.view", "reports.daily_sales", "reports.monthly_sales", "reports.best_selling", "reports.profit", "reports.stock", "reports.inventory_valuation", "reports.payment_summary", "reports.low_stock", "reports.reorder",
		"exports.view", "exports.inventory", "exports.products", "exports.sales", "exports.profit",
		"imports.view", "imports.template.download", "imports.products.preview", "imports.products.confirm", "imports.history.view",
		"suppliers.view", "suppliers.create", "suppliers.update", "suppliers.deactivate",
		"purchase_orders.view", "purchase_orders.create", "purchase_orders.update", "purchase_orders.send", "purchase_orders.receive", "purchase_orders.cancel", "purchase_orders.create_from_alert",
		"users.view", "users.create", "users.update", "users.deactivate", "users.assign_roles",
		"roles.view", "roles.create", "roles.update", "roles.deactivate", "roles.assign_permissions", "permissions.view",
		"settings.view", "settings.update", "settings.line.view", "settings.line.update", "settings.line.test",
		"notifications.view",
	}
	if role == RoleAdmin {
		return admin
	}
	manager := []string{
		"dashboard.view", "pos.view", "pos.sell", "products.view", "products.create", "products.update", "products.deactivate", "products.import", "products.export",
		"categories.view", "categories.create", "categories.update", "categories.deactivate",
		"stock.view", "stock.restock", "stock.adjust", "stock.movements.view",
		"locations.view", "locations.create", "locations.update", "locations.deactivate",
		"transfers.view", "transfers.create", "transfers.complete", "transfers.cancel",
		"sales.view", "sales.receipt.view", "sales.cancel", "alerts.view", "alerts.mark_read", "alerts.create_po",
		"reports.view", "reports.daily_sales", "reports.monthly_sales", "reports.best_selling", "reports.profit", "reports.stock", "reports.inventory_valuation", "reports.payment_summary", "reports.low_stock", "reports.reorder",
		"exports.view", "exports.inventory", "exports.products", "exports.sales", "exports.profit",
		"imports.view", "imports.template.download", "imports.products.preview", "imports.products.confirm", "imports.history.view",
		"suppliers.view", "suppliers.create", "suppliers.update", "suppliers.deactivate",
		"purchase_orders.view", "purchase_orders.create", "purchase_orders.update", "purchase_orders.send", "purchase_orders.receive", "purchase_orders.cancel", "purchase_orders.create_from_alert",
		"notifications.view",
	}
	if role == RoleManager {
		return manager
	}
	return []string{"dashboard.view", "pos.view", "pos.sell", "pos.clear_cart", "products.view", "sales.view", "sales.receipt.view", "alerts.view", "alerts.mark_read"}
}
