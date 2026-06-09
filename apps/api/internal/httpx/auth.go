package httpx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleManager Role = "MANAGER"
	RoleCashier Role = "CASHIER"
)

type User struct {
	ID              uint64     `json:"id"`
	Username        string     `json:"username"`
	FullName        string     `json:"fullName"`
	Role            Role       `json:"role"`
	Roles           []UserRole `json:"roles,omitempty"`
	Active          bool       `json:"active"`
	AvatarURL       string     `json:"avatar_url"`
	AvatarUpdatedAt *time.Time `json:"avatar_updated_at,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type UserRole struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type contextKey string

const userContextKey contextKey = "authUser"

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	user, hash, err := s.userByUsername(r.Context(), body.Username)
	if err != nil || !user.Active {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid username or password.")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid username or password.")
		return
	}

	token, err := s.issueToken(user)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "TOKEN_ERROR", "Could not create access token.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"token":       token,
		"user":        user,
		"roles":       s.rolesForUser(r.Context(), user),
		"permissions": s.permissionsForUser(r.Context(), user),
	})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"user":        user,
		"roles":       s.rolesForUser(r.Context(), user),
		"permissions": s.permissionsForUser(r.Context(), user),
	})
}

func (s *Server) users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.QueryContext(r.Context(), `SELECT id, username, full_name, role, active, COALESCE(avatar_url, ''), avatar_updated_at, created_at FROM users ORDER BY id`)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load users.")
			return
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var user User
			var avatarUpdatedAt sql.NullTime
			if err := rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Role, &user.Active, &user.AvatarURL, &avatarUpdatedAt, &user.CreatedAt); err != nil {
				response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read users.")
				return
			}
			if avatarUpdatedAt.Valid {
				user.AvatarUpdatedAt = &avatarUpdatedAt.Time
			}
			user.Roles = s.rolesForUser(r.Context(), user)
			users = append(users, user)
		}
		response.JSON(w, http.StatusOK, users)
	case http.MethodPost:
		var body userInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		user, err := s.createUser(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "CREATE_USER_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, user)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) userDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := s.userByID(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
			return
		}
		response.JSON(w, http.StatusOK, user)
	case http.MethodPatch:
		var body userInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if current, ok := currentUser(r.Context()); ok && current.ID == id && !body.Active {
			response.ErrorJSON(w, http.StatusBadRequest, "SELF_DEACTIVATION_FORBIDDEN", "You cannot deactivate your own account.")
			return
		}
		user, err := s.updateUser(r.Context(), id, body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_USER_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusOK, user)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) userStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id.")
		return
	}

	var body struct {
		Active bool `json:"active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if current, ok := currentUser(r.Context()); ok && current.ID == id && !body.Active {
		response.ErrorJSON(w, http.StatusBadRequest, "SELF_DEACTIVATION_FORBIDDEN", "You cannot deactivate your own account.")
		return
	}

	_, err = s.db.ExecContext(r.Context(), `UPDATE users SET active=? WHERE id=?`, body.Active, id)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_STATUS_FAILED", "Could not update user status.")
		return
	}
	user, err := s.userByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

type userInput struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	FullName string   `json:"fullName"`
	Role     Role     `json:"role"`
	RoleIDs  []uint64 `json:"role_ids"`
	Active   bool     `json:"active"`
}

func (s *Server) createUser(ctx context.Context, body userInput) (User, error) {
	if strings.TrimSpace(body.Username) == "" || strings.TrimSpace(body.FullName) == "" || body.Password == "" {
		return User{}, errors.New("username, full name, and password are required")
	}
	if !validRole(body.Role) {
		return User{}, errors.New("invalid role")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO users(username, password_hash, full_name, role, active) VALUES (?, ?, ?, ?, ?)`, body.Username, string(hash), body.FullName, body.Role, body.Active)
	if err != nil {
		return User{}, err
	}
	id, _ := result.LastInsertId()
	if err := s.assignRolesToUser(ctx, uint64(id), body.RoleIDs, body.Role); err != nil {
		return User{}, err
	}
	return s.userByID(ctx, uint64(id))
}

func (s *Server) updateUser(ctx context.Context, id uint64, body userInput) (User, error) {
	if strings.TrimSpace(body.Username) == "" || strings.TrimSpace(body.FullName) == "" {
		return User{}, errors.New("username and full name are required")
	}
	if !validRole(body.Role) {
		return User{}, errors.New("invalid role")
	}
	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		_, err = s.db.ExecContext(ctx, `UPDATE users SET username=?, password_hash=?, full_name=?, role=?, active=? WHERE id=?`, body.Username, string(hash), body.FullName, body.Role, body.Active, id)
		if err != nil {
			return User{}, err
		}
	} else {
		_, err := s.db.ExecContext(ctx, `UPDATE users SET username=?, full_name=?, role=?, active=? WHERE id=?`, body.Username, body.FullName, body.Role, body.Active, id)
		if err != nil {
			return User{}, err
		}
	}
	if err := s.assignRolesToUser(ctx, id, body.RoleIDs, body.Role); err != nil {
		return User{}, err
	}
	return s.userByID(ctx, id)
}

func (s *Server) userByUsername(ctx context.Context, username string) (User, string, error) {
	var user User
	var hash string
	var avatarUpdatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id, username, full_name, role, active, COALESCE(avatar_url, ''), avatar_updated_at, created_at, password_hash FROM users WHERE username=?`, username).
		Scan(&user.ID, &user.Username, &user.FullName, &user.Role, &user.Active, &user.AvatarURL, &avatarUpdatedAt, &user.CreatedAt, &hash)
	if avatarUpdatedAt.Valid {
		user.AvatarUpdatedAt = &avatarUpdatedAt.Time
	}
	return user, hash, err
}

func (s *Server) userByID(ctx context.Context, id uint64) (User, error) {
	var user User
	var avatarUpdatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id, username, full_name, role, active, COALESCE(avatar_url, ''), avatar_updated_at, created_at FROM users WHERE id=?`, id).
		Scan(&user.ID, &user.Username, &user.FullName, &user.Role, &user.Active, &user.AvatarURL, &avatarUpdatedAt, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}
	if avatarUpdatedAt.Valid {
		user.AvatarUpdatedAt = &avatarUpdatedAt.Time
	}
	user.Roles = s.rolesForUser(ctx, user)
	return user, err
}

func (s *Server) issueToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  strconv.FormatUint(user.ID, 10),
		"role": string(user.Role),
		"exp":  time.Now().Add(12 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
			return
		}

		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(token *jwt.Token) (any, error) {
			return []byte(s.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token.")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims.")
			return
		}
		sub, _ := claims["sub"].(string)
		id, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token subject.")
			return
		}

		user, err := s.userByID(r.Context(), id)
		if err != nil || !user.Active {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User is inactive or missing.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	}
}

func (s *Server) requireRoles(next http.HandlerFunc, roles ...Role) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		for _, role := range roles {
			if user.Role == role {
				next.ServeHTTP(w, r)
				return
			}
		}
		response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource.")
	})
}

func (s *Server) requirePermission(next http.HandlerFunc, permission string) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		if s.userHasPermission(r.Context(), user, permission) {
			next.ServeHTTP(w, r)
			return
		}
		response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource.")
	})
}

func (s *Server) requireAnyPermission(next http.HandlerFunc, permissions ...string) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		for _, permission := range permissions {
			if s.userHasPermission(r.Context(), user, permission) {
				next.ServeHTTP(w, r)
				return
			}
		}
		response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource.")
	})
}

func (s *Server) requireAllPermissions(next http.HandlerFunc, permissions ...string) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		for _, permission := range permissions {
			if !s.userHasPermission(r.Context(), user, permission) {
				response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource.")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

func validRole(role Role) bool {
	return role == RoleAdmin || role == RoleManager || role == RoleCashier
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func parsePathID(r *http.Request, key string) (uint64, error) {
	return strconv.ParseUint(r.PathValue(key), 10, 64)
}
