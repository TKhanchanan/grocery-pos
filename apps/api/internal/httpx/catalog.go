package httpx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type Category struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Location struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID           uint64         `json:"id"`
	SKU          string         `json:"sku"`
	Name         string         `json:"name"`
	Barcode      *string        `json:"barcode"`
	CategoryID   *uint64        `json:"category_id"`
	CategoryName *string        `json:"category_name"`
	SellingPrice float64        `json:"selling_price"`
	UnitCost     float64        `json:"unit_cost"`
	Unit         string         `json:"unit"`
	Threshold    int            `json:"threshold"`
	ReorderPoint int            `json:"reorder_point"`
	Active       bool           `json:"is_active"`
	ImageURL     *string        `json:"image_url"`
	ImageUpdated *time.Time     `json:"image_updated_at"`
	TotalStock   int            `json:"total_stock"`
	StockStatus  string         `json:"stock_status"`
	Stocks       []ProductStock `json:"stocks,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

const maxProductImageBytes = 2 * 1024 * 1024

type ProductStock struct {
	ProductID    uint64 `json:"product_id"`
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	LocationID   uint64 `json:"location_id"`
	LocationName string `json:"location_name"`
	Quantity     int    `json:"quantity"`
	StockStatus  string `json:"stock_status"`
}

type categoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type locationInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type productInput struct {
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Barcode      *string `json:"barcode"`
	CategoryID   *uint64 `json:"category_id"`
	SellingPrice float64 `json:"selling_price"`
	UnitCost     float64 `json:"unit_cost"`
	Unit         string  `json:"unit"`
	Threshold    int     `json:"threshold"`
	ReorderPoint int     `json:"reorder_point"`
	Active       bool    `json:"is_active"`
}

func (s *Server) categories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.QueryContext(r.Context(), `SELECT id, name, description, active, created_at FROM categories ORDER BY name`)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load categories.")
			return
		}
		defer rows.Close()
		categories := []Category{}
		for rows.Next() {
			var item Category
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt); err != nil {
				response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read categories.")
				return
			}
			categories = append(categories, item)
		}
		response.JSON(w, http.StatusOK, categories)
	case http.MethodPost:
		var body categoryInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		category, err := s.createCategory(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "CREATE_CATEGORY_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, category)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) categoryDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid category id.")
		return
	}
	var body categoryInput
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	category, err := s.updateCategory(r.Context(), id, body)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_CATEGORY_FAILED", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, category)
}

func (s *Server) categoryStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid category id.")
		return
	}
	var body struct {
		Active bool `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if !body.Active {
		var activeProducts int
		if err := s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM products WHERE category_id=? AND active=TRUE`, id).Scan(&activeProducts); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not check category products.")
			return
		}
		if activeProducts > 0 {
			response.ErrorJSON(w, http.StatusConflict, "CATEGORY_HAS_ACTIVE_PRODUCTS", "Category still has active products.")
			return
		}
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE categories SET active=? WHERE id=?`, body.Active, id); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_CATEGORY_STATUS_FAILED", "Could not update category status.")
		return
	}
	category, err := s.categoryByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Category not found.")
		return
	}
	response.JSON(w, http.StatusOK, category)
}

func (s *Server) products(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		products, err := s.listProducts(r.Context(), r)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load products.")
			return
		}
		response.JSON(w, http.StatusOK, products)
	case http.MethodPost:
		var body productInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		product, err := s.createProduct(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, productErrorStatus(err), productErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, product)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) productDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		product, err := s.productByID(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Product not found.")
			return
		}
		response.JSON(w, http.StatusOK, product)
	case http.MethodPatch:
		var body productInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		product, err := s.updateProduct(r.Context(), id, body)
		if err != nil {
			response.ErrorJSON(w, productErrorStatus(err), productErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusOK, product)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) productStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	var body struct {
		Active bool `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE products SET active=? WHERE id=?`, body.Active, id); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_PRODUCT_STATUS_FAILED", "Could not update product status.")
		return
	}
	product, err := s.productByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Product not found.")
		return
	}
	response.JSON(w, http.StatusOK, product)
}

func (s *Server) productImage(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	if _, err := s.productByID(r.Context(), id); err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Product not found.")
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.uploadProductImage(w, r, id)
	case http.MethodDelete:
		s.deleteProductImage(w, r, id)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) uploadProductImage(w http.ResponseWriter, r *http.Request, id uint64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProductImageBytes+1024)
	if err := r.ParseMultipartForm(maxProductImageBytes); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "PRODUCT_IMAGE_TOO_LARGE", "Product image must be 2MB or smaller.")
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "PRODUCT_IMAGE_REQUIRED", "Choose a product image to upload.")
		return
	}
	defer file.Close()

	mime, ext, err := detectProductImageType(file)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_PRODUCT_IMAGE_TYPE", err.Error())
		return
	}
	if header.Size > maxProductImageBytes {
		response.ErrorJSON(w, http.StatusBadRequest, "PRODUCT_IMAGE_TOO_LARGE", "Product image must be 2MB or smaller.")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not read product image.")
		return
	}

	dir := filepath.Join(s.cfg.UploadDir, "products")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not prepare upload storage.")
		return
	}
	name := fmt.Sprintf("%d-%s%s", id, randomHex(12), ext)
	path := filepath.Join(dir, name)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not save product image.")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		_ = out.Close()
		_ = os.Remove(path)
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not save product image.")
		return
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(path)
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not finish saving product image.")
		return
	}

	oldPath := productImagePath(r.Context(), s.db, id)
	imageURL := "/uploads/products/" + name
	if _, err := s.db.ExecContext(r.Context(), `UPDATE products SET image_url=?, image_path=?, image_updated_at=? WHERE id=?`, imageURL, path, time.Now(), id); err != nil {
		_ = os.Remove(path)
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_UPLOAD_FAILED", "Could not update product image.")
		return
	}
	removeAvatarFile(oldPath)
	product, _ := s.productByID(r.Context(), id)
	response.JSON(w, http.StatusOK, map[string]any{
		"product":      product,
		"content_type": mime,
	})
}

func (s *Server) deleteProductImage(w http.ResponseWriter, r *http.Request, id uint64) {
	oldPath := productImagePath(r.Context(), s.db, id)
	if _, err := s.db.ExecContext(r.Context(), `UPDATE products SET image_url=NULL, image_path=NULL, image_updated_at=NULL WHERE id=?`, id); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "PRODUCT_IMAGE_DELETE_FAILED", "Could not remove product image.")
		return
	}
	removeAvatarFile(oldPath)
	product, _ := s.productByID(r.Context(), id)
	response.JSON(w, http.StatusOK, product)
}

func detectProductImageType(file multipartSeekReader) (string, string, error) {
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
		return "", "", errors.New("Only JPG, PNG, and WEBP product images are supported.")
	}
}

func productImagePath(ctx context.Context, db *sql.DB, productID uint64) string {
	var path sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT image_path FROM products WHERE id=?`, productID).Scan(&path)
	if path.Valid {
		return path.String
	}
	return ""
}

func (s *Server) locations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.QueryContext(r.Context(), `SELECT id, name, description, active, created_at FROM locations ORDER BY name`)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load locations.")
			return
		}
		defer rows.Close()
		locations := []Location{}
		for rows.Next() {
			var item Location
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt); err != nil {
				response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read locations.")
				return
			}
			locations = append(locations, item)
		}
		response.JSON(w, http.StatusOK, locations)
	case http.MethodPost:
		var body locationInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		location, err := s.createLocation(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "CREATE_LOCATION_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, location)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) locationDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid location id.")
		return
	}
	var body locationInput
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	location, err := s.updateLocation(r.Context(), id, body)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_LOCATION_FAILED", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, location)
}

func (s *Server) locationStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid location id.")
		return
	}
	var body struct {
		Active bool `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE locations SET active=? WHERE id=?`, body.Active, id); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "UPDATE_LOCATION_STATUS_FAILED", "Could not update location status.")
		return
	}
	location, err := s.locationByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Location not found.")
		return
	}
	response.JSON(w, http.StatusOK, location)
}

func (s *Server) productStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := s.queryProductStocks(r.Context(), nil)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load product stocks.")
		return
	}
	response.JSON(w, http.StatusOK, stocks)
}

func (s *Server) productStocksByProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	stocks, err := s.queryProductStocks(r.Context(), &id)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load product stocks.")
		return
	}
	response.JSON(w, http.StatusOK, stocks)
}

func (s *Server) createCategory(ctx context.Context, body categoryInput) (Category, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Category{}, errors.New("category name is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO categories(name, description, active) VALUES (?, ?, TRUE)`, body.Name, body.Description)
	if err != nil {
		return Category{}, err
	}
	id, _ := result.LastInsertId()
	return s.categoryByID(ctx, uint64(id))
}

func (s *Server) updateCategory(ctx context.Context, id uint64, body categoryInput) (Category, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Category{}, errors.New("category name is required")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE categories SET name=?, description=? WHERE id=?`, body.Name, body.Description, id); err != nil {
		return Category{}, err
	}
	return s.categoryByID(ctx, id)
}

func (s *Server) categoryByID(ctx context.Context, id uint64) (Category, error) {
	var item Category
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, active, created_at FROM categories WHERE id=?`, id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt)
	return item, err
}

func (s *Server) createLocation(ctx context.Context, body locationInput) (Location, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Location{}, errors.New("location name is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO locations(name, description, active) VALUES (?, ?, TRUE)`, body.Name, body.Description)
	if err != nil {
		return Location{}, err
	}
	id, _ := result.LastInsertId()
	return s.locationByID(ctx, uint64(id))
}

func (s *Server) updateLocation(ctx context.Context, id uint64, body locationInput) (Location, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Location{}, errors.New("location name is required")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE locations SET name=?, description=? WHERE id=?`, body.Name, body.Description, id); err != nil {
		return Location{}, err
	}
	return s.locationByID(ctx, id)
}

func (s *Server) locationByID(ctx context.Context, id uint64) (Location, error) {
	var item Location
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, active, created_at FROM locations WHERE id=?`, id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt)
	return item, err
}

func (s *Server) listProducts(ctx context.Context, r *http.Request) ([]Product, error) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	categoryID := strings.TrimSpace(r.URL.Query().Get("category_id"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	stockStatus := strings.TrimSpace(r.URL.Query().Get("stock_status"))

	where := []string{"1=1"}
	args := []any{}
	if query != "" {
		where = append(where, "(p.name LIKE ? OR p.sku LIKE ? OR p.barcode LIKE ?)")
		like := "%" + query + "%"
		args = append(args, like, like, like)
	}
	if categoryID != "" {
		where = append(where, "p.category_id=?")
		args = append(args, categoryID)
	}
	if status == "active" {
		where = append(where, "p.active=TRUE")
	}
	if status == "inactive" {
		where = append(where, "p.active=FALSE")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.sku, p.name, p.barcode, p.category_id, c.name, p.price, p.cost,
		       p.unit, p.threshold, p.reorder_point, p.active, p.image_url, p.image_updated_at,
		       COALESCE(SUM(ps.quantity), 0), p.created_at
		FROM products p
		LEFT JOIN categories c ON c.id=p.category_id
		LEFT JOIN product_stocks ps ON ps.product_id=p.id
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY p.id, c.name
		ORDER BY p.name, p.sku`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		if stockStatus == "" || product.StockStatus == stockStatus {
			products = append(products, product)
		}
	}
	return products, rows.Err()
}

func (s *Server) createProduct(ctx context.Context, body productInput) (Product, error) {
	if err := validateProduct(body); err != nil {
		return Product{}, err
	}
	if err := s.validateActiveCategory(ctx, body.CategoryID); err != nil {
		return Product{}, err
	}
	barcode := nullableBarcode(body.Barcode)
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO products(category_id, sku, barcode, name, unit, price, cost, threshold, reorder_point, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		body.CategoryID, body.SKU, barcode, body.Name, productUnit(body.Unit), body.SellingPrice, body.UnitCost, positiveOrZero(body.Threshold), positiveOrZero(body.ReorderPoint), body.Active)
	if err != nil {
		return Product{}, normalizeDuplicateError(err)
	}
	id, _ := result.LastInsertId()
	if err := s.ensureProductStockRows(ctx, uint64(id)); err != nil {
		return Product{}, err
	}
	return s.productByID(ctx, uint64(id))
}

func (s *Server) updateProduct(ctx context.Context, id uint64, body productInput) (Product, error) {
	if err := validateProduct(body); err != nil {
		return Product{}, err
	}
	if err := s.validateActiveCategory(ctx, body.CategoryID); err != nil {
		return Product{}, err
	}
	barcode := nullableBarcode(body.Barcode)
	_, err := s.db.ExecContext(ctx, `
		UPDATE products SET category_id=?, sku=?, barcode=?, name=?, unit=?, price=?, cost=?, threshold=?, reorder_point=?, active=?
		WHERE id=?`,
		body.CategoryID, body.SKU, barcode, body.Name, productUnit(body.Unit), body.SellingPrice, body.UnitCost, positiveOrZero(body.Threshold), positiveOrZero(body.ReorderPoint), body.Active, id)
	if err != nil {
		return Product{}, normalizeDuplicateError(err)
	}
	if err := s.ensureProductStockRows(ctx, id); err != nil {
		return Product{}, err
	}
	return s.productByID(ctx, id)
}

func (s *Server) validateActiveCategory(ctx context.Context, categoryID *uint64) error {
	if categoryID == nil {
		return nil
	}
	var active bool
	if err := s.db.QueryRowContext(ctx, `SELECT active FROM categories WHERE id=?`, *categoryID).Scan(&active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("category not found")
		}
		return err
	}
	if !active {
		return errors.New("category is inactive")
	}
	return nil
}

func (s *Server) productByID(ctx context.Context, id uint64) (Product, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.sku, p.name, p.barcode, p.category_id, c.name, p.price, p.cost,
		       p.unit, p.threshold, p.reorder_point, p.active, p.image_url, p.image_updated_at,
		       COALESCE(SUM(ps.quantity), 0), p.created_at
		FROM products p
		LEFT JOIN categories c ON c.id=p.category_id
		LEFT JOIN product_stocks ps ON ps.product_id=p.id
		WHERE p.id=?
		GROUP BY p.id, c.name`, id)
	if err != nil {
		return Product{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return Product{}, sql.ErrNoRows
	}
	product, err := scanProduct(rows)
	if err != nil {
		return Product{}, err
	}
	product.Stocks, _ = s.queryProductStocks(ctx, &id)
	return product, rows.Err()
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanProduct(scanner productScanner) (Product, error) {
	var product Product
	var imageURL sql.NullString
	var imageUpdated sql.NullTime
	if err := scanner.Scan(&product.ID, &product.SKU, &product.Name, &product.Barcode, &product.CategoryID, &product.CategoryName, &product.SellingPrice, &product.UnitCost, &product.Unit, &product.Threshold, &product.ReorderPoint, &product.Active, &imageURL, &imageUpdated, &product.TotalStock, &product.CreatedAt); err != nil {
		return Product{}, err
	}
	if imageURL.Valid {
		product.ImageURL = &imageURL.String
	}
	if imageUpdated.Valid {
		product.ImageUpdated = &imageUpdated.Time
	}
	product.StockStatus = stockStatus(product.TotalStock, product.Threshold, product.ReorderPoint)
	return product, nil
}

func (s *Server) queryProductStocks(ctx context.Context, productID *uint64) ([]ProductStock, error) {
	where := "1=1"
	args := []any{}
	if productID != nil {
		where = "p.id=?"
		args = append(args, *productID)
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.sku, l.id, l.name, COALESCE(ps.quantity, 0), p.threshold, p.reorder_point
		FROM products p
		CROSS JOIN locations l
		LEFT JOIN product_stocks ps ON ps.product_id=p.id AND ps.location_id=l.id
		WHERE `+where+`
		ORDER BY p.name, l.name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stocks := []ProductStock{}
	for rows.Next() {
		var item ProductStock
		var threshold, reorderPoint int
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.SKU, &item.LocationID, &item.LocationName, &item.Quantity, &threshold, &reorderPoint); err != nil {
			return nil, err
		}
		item.StockStatus = stockStatus(item.Quantity, threshold, reorderPoint)
		stocks = append(stocks, item)
	}
	return stocks, rows.Err()
}

func (s *Server) ensureProductStockRows(ctx context.Context, productID uint64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT IGNORE INTO product_stocks(product_id, location_id, quantity)
		SELECT ?, id, 0 FROM locations`, productID)
	return err
}

func validateProduct(body productInput) error {
	if strings.TrimSpace(body.SKU) == "" {
		return errors.New("sku is required")
	}
	if strings.TrimSpace(body.Name) == "" {
		return errors.New("product name is required")
	}
	if body.SellingPrice <= 0 {
		return errors.New("price must be greater than 0")
	}
	if body.UnitCost < 0 {
		return errors.New("unit cost must be greater than or equal to 0")
	}
	return nil
}

func productUnit(unit string) string {
	if strings.TrimSpace(unit) == "" {
		return "ชิ้น"
	}
	return strings.TrimSpace(unit)
}

func nullableBarcode(barcode *string) any {
	if barcode == nil || strings.TrimSpace(*barcode) == "" {
		return nil
	}
	return strings.TrimSpace(*barcode)
}

func positiveOrZero(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeDuplicateError(err error) error {
	message := err.Error()
	if strings.Contains(message, "Duplicate entry") && strings.Contains(message, "sku") {
		return fmt.Errorf("SKU already exists")
	}
	if strings.Contains(message, "Duplicate entry") && strings.Contains(message, "barcode") {
		return fmt.Errorf("barcode already exists")
	}
	return err
}

func productErrorCode(err error) string {
	if strings.Contains(err.Error(), "already exists") {
		return "DUPLICATE_VALUE"
	}
	return "PRODUCT_VALIDATION_FAILED"
}

func productErrorStatus(err error) int {
	if strings.Contains(err.Error(), "already exists") {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}

func stockStatus(quantity int, threshold int, reorderPoint int) string {
	if quantity == 0 {
		return "out_of_stock"
	}
	if threshold > 0 && quantity <= threshold {
		return "low_stock"
	}
	if reorderPoint > 0 && quantity <= reorderPoint {
		return "reorder_point"
	}
	return "in_stock"
}
