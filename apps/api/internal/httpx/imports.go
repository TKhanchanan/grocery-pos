package httpx

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type ProductImportRow struct {
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Barcode      string  `json:"barcode"`
	Category     string  `json:"category"`
	SellingPrice float64 `json:"selling_price"`
	UnitCost     float64 `json:"unit_cost"`
	Threshold    int     `json:"threshold"`
	ReorderPoint int     `json:"reorder_point"`
	Location     string  `json:"location"`
	InitialStock *int    `json:"initial_stock"`
}

type ImportJob struct {
	ID          uint64         `json:"id"`
	JobType     string         `json:"job_type"`
	FileName    string         `json:"file_name"`
	Status      string         `json:"status"`
	TotalRows   int            `json:"total_rows"`
	SuccessRows int            `json:"success_rows"`
	FailedRows  int            `json:"failed_rows"`
	CreatedBy   *uint64        `json:"created_by"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	Rows        []ImportJobRow `json:"rows,omitempty"`
}

type ImportJobRow struct {
	ID           uint64           `json:"id"`
	ImportJobID  uint64           `json:"import_job_id"`
	RowIndex     int              `json:"row_index"`
	RawData      ProductImportRow `json:"raw_data"`
	Status       string           `json:"status"`
	ErrorMessage string           `json:"error_message"`
	CreatedAt    time.Time        `json:"created_at"`
}

func (s *Server) productImportTemplate(w http.ResponseWriter, r *http.Request) {
	rows := [][]string{
		{"sku", "name", "barcode", "category", "selling_price", "unit_cost", "threshold", "reorder_point", "location", "initial_stock"},
		{"RICE-001", "ข้าวหอมมะลิ", "885000000101", "อาหาร", "120.00", "80.00", "5", "10", "หน้าร้าน", "20"},
	}
	writeCSV(w, "product-import-template.csv", rows)
}

func (s *Server) productImportPreview(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context())
	file, header, err := importUploadFile(r)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	defer file.Close()

	rows, err := parseProductImportCSV(file, header.Filename)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	job, err := s.createProductImportPreview(r.Context(), user, header.Filename, rows)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "IMPORT_PREVIEW_FAILED", "Could not create import preview.")
		return
	}
	response.JSON(w, http.StatusCreated, job)
}

func (s *Server) productImportConfirm(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JobID uint64 `json:"job_id"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if body.JobID == 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "job_id is required.")
		return
	}
	user, _ := currentUser(r.Context())
	job, err := s.confirmProductImport(r.Context(), user, body.JobID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusBadRequest, "IMPORT_CONFIRM_FAILED", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, job)
}

func (s *Server) imports(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.listImportJobs(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load imports.")
		return
	}
	response.JSON(w, http.StatusOK, jobs)
}

func (s *Server) importDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid import id.")
		return
	}
	job, err := s.importJobByID(r.Context(), id, true)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Import job not found.")
		return
	}
	response.JSON(w, http.StatusOK, job)
}

func importUploadFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, nil, errors.New("multipart file upload is required")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, errors.New("file field is required")
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" {
		file.Close()
		return nil, nil, errors.New("only csv product imports are supported")
	}
	return file, header, nil
}

func parseProductImportCSV(file io.Reader, fileName string) ([]ProductImportRow, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("csv requires a header row and at least one product row")
	}
	header := map[string]int{}
	for index, name := range records[0] {
		header[strings.TrimSpace(strings.ToLower(strings.TrimPrefix(name, "\uFEFF")))] = index
	}
	required := []string{"sku", "name", "barcode", "category", "selling_price", "unit_cost", "threshold", "reorder_point", "location", "initial_stock"}
	for _, field := range required {
		if _, ok := header[field]; !ok {
			return nil, fmt.Errorf("missing required template column %s", field)
		}
	}
	rows := []ProductImportRow{}
	for _, record := range records[1:] {
		row := ProductImportRow{
			SKU:      csvValue(record, header, "sku"),
			Name:     csvValue(record, header, "name"),
			Barcode:  csvValue(record, header, "barcode"),
			Category: csvValue(record, header, "category"),
			Location: csvValue(record, header, "location"),
		}
		row.SellingPrice, _ = strconv.ParseFloat(blankZero(csvValue(record, header, "selling_price")), 64)
		row.UnitCost, _ = strconv.ParseFloat(blankZero(csvValue(record, header, "unit_cost")), 64)
		row.Threshold, _ = strconv.Atoi(blankZero(csvValue(record, header, "threshold")))
		row.ReorderPoint, _ = strconv.Atoi(blankZero(csvValue(record, header, "reorder_point")))
		if value := strings.TrimSpace(csvValue(record, header, "initial_stock")); value != "" {
			stock, _ := strconv.Atoi(value)
			row.InitialStock = &stock
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func csvValue(record []string, header map[string]int, name string) string {
	index := header[name]
	if index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func blankZero(value string) string {
	if strings.TrimSpace(value) == "" {
		return "0"
	}
	return strings.TrimSpace(value)
}

func (s *Server) createProductImportPreview(ctx context.Context, user User, fileName string, rows []ProductImportRow) (ImportJob, error) {
	seenSKUs := map[string]bool{}
	seenBarcodes := map[string]bool{}
	var jobID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO import_jobs(job_type, file_name, status, total_rows, created_by)
			VALUES ('PRODUCTS', ?, 'PENDING', ?, ?)`, fileName, len(rows), user.ID)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		jobID = uint64(id)
		success, failed := 0, 0
		for index, row := range rows {
			status := "PENDING"
			message := s.validateImportRow(ctx, tx, row, seenSKUs, seenBarcodes)
			if message != "" {
				status = "FAILED"
				failed++
			} else {
				success++
			}
			raw, _ := json.Marshal(row)
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO import_job_rows(import_job_id, row_index, raw_data, status, error_message)
				VALUES (?, ?, ?, ?, ?)`, jobID, index+2, string(raw), status, message); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE import_jobs SET success_rows=?, failed_rows=? WHERE id=?`, success, failed, jobID)
		return err
	})
	if err != nil {
		return ImportJob{}, err
	}
	return s.importJobByID(ctx, jobID, true)
}

func (s *Server) validateImportRow(ctx context.Context, tx *sql.Tx, row ProductImportRow, seenSKUs, seenBarcodes map[string]bool) string {
	errors := []string{}
	sku := strings.TrimSpace(row.SKU)
	if sku == "" {
		errors = append(errors, "SKU is required")
	} else {
		skuKey := strings.ToLower(sku)
		if seenSKUs[skuKey] {
			errors = append(errors, "duplicate SKU in file")
		}
		seenSKUs[skuKey] = true
		var exists int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE sku=?`, sku).Scan(&exists)
		if exists > 0 {
			errors = append(errors, "duplicate SKU already exists")
		}
	}
	if strings.TrimSpace(row.Name) == "" {
		errors = append(errors, "Name is required")
	}
	if row.SellingPrice <= 0 {
		errors = append(errors, "Selling price must be greater than 0")
	}
	if row.UnitCost < 0 {
		errors = append(errors, "Unit cost must be greater than or equal to 0")
	}
	if row.Threshold < 0 {
		errors = append(errors, "Threshold must be greater than or equal to 0")
	}
	if row.ReorderPoint < 0 {
		errors = append(errors, "Reorder point must be greater than or equal to 0")
	}
	if row.InitialStock != nil && *row.InitialStock < 0 {
		errors = append(errors, "Initial stock must be greater than or equal to 0")
	}
	barcode := strings.TrimSpace(row.Barcode)
	if barcode != "" {
		if seenBarcodes[barcode] {
			errors = append(errors, "duplicate barcode in file")
		}
		seenBarcodes[barcode] = true
		var exists int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE barcode=?`, barcode).Scan(&exists)
		if exists > 0 {
			errors = append(errors, "barcode already exists")
		}
	}
	return strings.Join(errors, "; ")
}

func (s *Server) confirmProductImport(ctx context.Context, user User, jobID uint64) (ImportJob, error) {
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM import_jobs WHERE id=? AND job_type='PRODUCTS' FOR UPDATE`, jobID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("import job not found")
			}
			return err
		}
		if status == "COMPLETED" {
			return errors.New("import job is already completed")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE import_jobs SET status='PROCESSING', started_at=NOW() WHERE id=?`, jobID); err != nil {
			return err
		}
		rows, err := importRowsForConfirm(ctx, tx, jobID)
		if err != nil {
			return err
		}
		imported := 0
		for _, row := range rows {
			if _, err := s.insertImportedProduct(ctx, tx, user, row.RawData); err != nil {
				_, _ = tx.ExecContext(ctx, `UPDATE import_job_rows SET status='FAILED', error_message=? WHERE id=?`, err.Error(), row.ID)
				continue
			}
			imported++
			_, _ = tx.ExecContext(ctx, `UPDATE import_job_rows SET status='IMPORTED', error_message='' WHERE id=?`, row.ID)
		}
		var failed int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM import_job_rows WHERE import_job_id=? AND status='FAILED'`, jobID).Scan(&failed); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE import_jobs SET status='COMPLETED', success_rows=?, failed_rows=?, completed_at=NOW() WHERE id=?`, imported, failed, jobID)
		return err
	})
	if err != nil {
		return ImportJob{}, err
	}
	return s.importJobByID(ctx, jobID, true)
}

func importRowsForConfirm(ctx context.Context, tx *sql.Tx, jobID uint64) ([]ImportJobRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, import_job_id, row_index, raw_data, status, error_message, created_at
		FROM import_job_rows
		WHERE import_job_id=? AND status='PENDING'
		ORDER BY row_index`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ImportJobRow{}
	for rows.Next() {
		item, err := scanImportRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Server) insertImportedProduct(ctx context.Context, tx *sql.Tx, user User, row ProductImportRow) (uint64, error) {
	categoryID, err := findOrCreateCategory(ctx, tx, row.Category)
	if err != nil {
		return 0, err
	}
	barcode := any(nil)
	if strings.TrimSpace(row.Barcode) != "" {
		barcode = strings.TrimSpace(row.Barcode)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO products(category_id, sku, barcode, name, unit, price, cost, threshold, reorder_point, active)
		VALUES (?, ?, ?, ?, 'ชิ้น', ?, ?, ?, ?, TRUE)`,
		categoryID, row.SKU, barcode, row.Name, row.SellingPrice, row.UnitCost, positiveOrZero(row.Threshold), positiveOrZero(row.ReorderPoint))
	if err != nil {
		return 0, normalizeDuplicateError(err)
	}
	id, _ := result.LastInsertId()
	productID := uint64(id)
	if row.InitialStock != nil && *row.InitialStock > 0 {
		locationID, err := findOrCreateLocation(ctx, tx, row.Location)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO product_stocks(product_id, location_id, quantity) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE quantity=quantity+VALUES(quantity)`, productID, locationID, *row.InitialStock); err != nil {
			return 0, err
		}
		if _, err := insertStockMovement(ctx, tx, productID, locationID, "IMPORT", *row.InitialStock, 0, *row.InitialStock, &row.UnitCost, "product import", user.ID); err != nil {
			return 0, err
		}
		if err := recalculateAlerts(ctx, tx, productID, locationID); err != nil {
			return 0, err
		}
	}
	if err := ensureProductStockRowsTx(ctx, tx, productID); err != nil {
		return 0, err
	}
	return productID, nil
}

func ensureProductStockRowsTx(ctx context.Context, tx *sql.Tx, productID uint64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT IGNORE INTO product_stocks(product_id, location_id, quantity)
		SELECT ?, id, 0 FROM locations`, productID)
	return err
}

func findOrCreateCategory(ctx context.Context, tx *sql.Tx, name string) (*uint64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	var id uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM categories WHERE name=?`, name).Scan(&id); err == nil {
		return &id, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO categories(name, description, active) VALUES (?, '', TRUE)`, name)
	if err != nil {
		return nil, err
	}
	newID, _ := result.LastInsertId()
	id = uint64(newID)
	return &id, nil
}

func findOrCreateLocation(ctx context.Context, tx *sql.Tx, name string) (uint64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "หน้าร้าน"
	}
	var id uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM locations WHERE name=?`, name).Scan(&id); err == nil {
		return id, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO locations(name, description, active) VALUES (?, '', TRUE)`, name)
	if err != nil {
		return 0, err
	}
	newID, _ := result.LastInsertId()
	return uint64(newID), nil
}

func (s *Server) listImportJobs(ctx context.Context) ([]ImportJob, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, job_type, file_name, status, total_rows, success_rows, failed_rows, created_by, started_at, completed_at, created_at
		FROM import_jobs
		ORDER BY id DESC
		LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []ImportJob{}
	for rows.Next() {
		job, err := scanImportJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (s *Server) importJobByID(ctx context.Context, id uint64, includeRows bool) (ImportJob, error) {
	job, err := scanImportJob(s.db.QueryRowContext(ctx, `
		SELECT id, job_type, file_name, status, total_rows, success_rows, failed_rows, created_by, started_at, completed_at, created_at
		FROM import_jobs
		WHERE id=?`, id))
	if err != nil {
		return ImportJob{}, err
	}
	if includeRows {
		job.Rows, err = s.importJobRows(ctx, id)
	}
	return job, err
}

func (s *Server) importJobRows(ctx context.Context, jobID uint64) ([]ImportJobRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, import_job_id, row_index, raw_data, status, error_message, created_at
		FROM import_job_rows
		WHERE import_job_id=?
		ORDER BY row_index`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ImportJobRow{}
	for rows.Next() {
		item, err := scanImportRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type importJobScanner interface {
	Scan(dest ...any) error
}

func scanImportJob(scanner importJobScanner) (ImportJob, error) {
	var job ImportJob
	err := scanner.Scan(&job.ID, &job.JobType, &job.FileName, &job.Status, &job.TotalRows, &job.SuccessRows, &job.FailedRows, &job.CreatedBy, &job.StartedAt, &job.CompletedAt, &job.CreatedAt)
	return job, err
}

func scanImportRow(scanner importJobScanner) (ImportJobRow, error) {
	var item ImportJobRow
	var raw string
	if err := scanner.Scan(&item.ID, &item.ImportJobID, &item.RowIndex, &raw, &item.Status, &item.ErrorMessage, &item.CreatedAt); err != nil {
		return ImportJobRow{}, err
	}
	_ = json.Unmarshal([]byte(raw), &item.RawData)
	return item, nil
}
