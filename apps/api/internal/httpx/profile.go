package httpx

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

const maxAvatarBytes = 2 * 1024 * 1024

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		response.JSON(w, http.StatusOK, map[string]any{
			"user":        user,
			"roles":       s.rolesForUser(r.Context(), user),
			"permissions": s.permissionsForUser(r.Context(), user),
		})
	case http.MethodPatch:
		var body struct {
			FullName string `json:"fullName"`
		}
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if strings.TrimSpace(body.FullName) == "" {
			response.ErrorJSON(w, http.StatusBadRequest, "PROFILE_UPDATE_FAILED", "Full name is required.")
			return
		}
		if _, err := s.db.ExecContext(r.Context(), `UPDATE users SET full_name=? WHERE id=?`, strings.TrimSpace(body.FullName), user.ID); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "PROFILE_UPDATE_FAILED", "Could not update profile.")
			return
		}
		fresh, _ := s.userByID(r.Context(), user.ID)
		response.JSON(w, http.StatusOK, map[string]any{
			"user":        fresh,
			"roles":       s.rolesForUser(r.Context(), fresh),
			"permissions": s.permissionsForUser(r.Context(), fresh),
		})
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBytes+1024)
	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "AVATAR_TOO_LARGE", "Profile image must be 2MB or smaller.")
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "AVATAR_REQUIRED", "Choose a profile image to upload.")
		return
	}
	defer file.Close()

	mime, ext, err := detectAvatarType(file)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_AVATAR_TYPE", err.Error())
		return
	}
	if header.Size > maxAvatarBytes {
		response.ErrorJSON(w, http.StatusBadRequest, "AVATAR_TOO_LARGE", "Profile image must be 2MB or smaller.")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_UPLOAD_FAILED", "Could not read profile image.")
		return
	}

	dir := filepath.Join(s.cfg.UploadDir, "avatars")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_UPLOAD_FAILED", "Could not prepare upload storage.")
		return
	}
	name := randomHex(16) + ext
	path := filepath.Join(dir, name)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_UPLOAD_FAILED", "Could not save profile image.")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		_ = out.Close()
		_ = os.Remove(path)
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_UPLOAD_FAILED", "Could not save profile image.")
		return
	}
	_ = out.Close()

	oldPath := userAvatarPath(r.Context(), s.db, user.ID)
	avatarURL := "/uploads/avatars/" + name
	if _, err := s.db.ExecContext(r.Context(), `UPDATE users SET avatar_url=?, avatar_path=?, avatar_updated_at=? WHERE id=?`, avatarURL, path, time.Now(), user.ID); err != nil {
		_ = os.Remove(path)
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_UPLOAD_FAILED", "Could not update profile image.")
		return
	}
	removeAvatarFile(oldPath)

	fresh, _ := s.userByID(r.Context(), user.ID)
	response.JSON(w, http.StatusOK, map[string]any{
		"user":         fresh,
		"content_type": mime,
	})
}

func (s *Server) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	oldPath := userAvatarPath(r.Context(), s.db, user.ID)
	if _, err := s.db.ExecContext(r.Context(), `UPDATE users SET avatar_url=NULL, avatar_path=NULL, avatar_updated_at=NULL WHERE id=?`, user.ID); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "AVATAR_DELETE_FAILED", "Could not remove profile image.")
		return
	}
	removeAvatarFile(oldPath)
	fresh, _ := s.userByID(r.Context(), user.ID)
	response.JSON(w, http.StatusOK, fresh)
}

func detectAvatarType(file multipartSeekReader) (string, string, error) {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])
	switch mime {
	case "image/jpeg":
		return mime, ".jpg", nil
	case "image/png":
		return mime, ".png", nil
	case "image/webp":
		return mime, ".webp", nil
	default:
		return "", "", errors.New("Only JPG, PNG, and WEBP profile images are supported.")
	}
}

type multipartSeekReader interface {
	io.Reader
	io.Seeker
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(buf)
}

func userAvatarPath(ctx context.Context, db *sql.DB, userID uint64) string {
	var path sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT avatar_path FROM users WHERE id=?`, userID).Scan(&path)
	if path.Valid {
		return path.String
	}
	return ""
}

func removeAvatarFile(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
