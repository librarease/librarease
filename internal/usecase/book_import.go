package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
	"golang.org/x/sync/errgroup"
)

type PreviewImportBooksSummary struct {
	CreatedCount int
	UpdatedCount int
	InvalidCount int
}
type PreviewImportBooksRow struct {
	ID     *string
	Code   string
	Title  string
	Author string
	Status string
	Error  *string
}
type PreviewImportBooksResult struct {
	Path    string
	Summary PreviewImportBooksSummary
	Rows    []PreviewImportBooksRow
}

type ValidatedBookRow struct {
	RowNum int
	ID     *uuid.UUID
	Code   string
	Title  string
	Author string
	Year   int
	Status string // "create", "update", "invalid"
	ErrMsg string
}

func (u Usecase) validateImportBooksCSV(ctx context.Context, libID uuid.UUID, r io.Reader) ([]ValidatedBookRow, error) {
	type csvRow struct {
		rowNum int
		id     string
		code   string
		title  string
		author string
		year   int
	}

	csvChan := make(chan csvRow, 10)
	validatedChan := make(chan ValidatedBookRow, 10)
	g, ctx := errgroup.WithContext(ctx)

	// Stage 1: CSV Reader
	g.Go(func() error {
		defer close(csvChan)
		csvReader := csv.NewReader(r)

		header, err := csvReader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV header: %w", err)
		}
		if len(header) < 5 {
			return fmt.Errorf("invalid CSV format: expected columns (id, code, title, author, year)")
		}

		rowNum := 1
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read CSV row %d: %w", rowNum, err)
			}
			rowNum++

			var year int
			if record[4] != "" {
				fmt.Sscanf(record[4], "%d", &year)
			}

			select {
			case csvChan <- csvRow{
				rowNum: rowNum, id: record[0], code: record[1],
				title: record[2], author: record[3], year: year,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	// Stage 2: Validate against DB
	g.Go(func() error {
		defer close(validatedChan)

		existingBooks, _, err := u.repo.ListBooks(ctx, ListBooksOption{
			LibraryIDs: uuid.UUIDs{libID},
		})
		if err != nil {
			return fmt.Errorf("list books error: %w", err)
		}

		existingByCode := make(map[string]Book, len(existingBooks))
		existingByID := make(map[uuid.UUID]Book, len(existingBooks))
		for _, b := range existingBooks {
			existingByCode[b.Code] = b
			existingByID[b.ID] = b
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case row, ok := <-csvChan:
				if !ok {
					return nil
				}

				var id *uuid.UUID
				if row.id != "" {
					parsedID, err := uuid.Parse(row.id)
					if err != nil {
						v := ValidatedBookRow{
							RowNum: row.rowNum, Code: row.code,
							Title: row.title, Author: row.author, Year: row.year,
							Status: "invalid", ErrMsg: "invalid UUID",
						}
						validatedChan <- v
						continue
					}
					id = &parsedID
				}

				v := ValidatedBookRow{
					RowNum: row.rowNum, ID: id, Code: row.code,
					Title: row.title, Author: row.author, Year: row.year,
				}

				if row.title == "" || row.author == "" {
					v.Status, v.ErrMsg = "invalid", "missing required fields: title or author"
					validatedChan <- v
					continue
				}

				if row.id != "" {
					id, err := uuid.Parse(row.id)
					if err != nil {
						v.Status, v.ErrMsg = "invalid", "invalid UUID"
						validatedChan <- v
						continue
					}

					book, ok := existingByID[id]
					if !ok {
						v.Status, v.ErrMsg = "invalid", "book ID not found"
						validatedChan <- v
						continue
					}
					if book.LibraryID != libID {
						v.Status, v.ErrMsg = "invalid", "book not in your library"
						validatedChan <- v
						continue
					}

					if b := existingByID[id]; b.Title == v.Title && b.Author == v.Author && b.Year == v.Year && b.Code == v.Code {
						v.Status, v.ErrMsg = "invalid", "no changes detected"
						validatedChan <- v
						continue
					}

					v.Status, v.ID = "update", &id
					validatedChan <- v
					continue
				}

				if row.code == "" {
					v.Status, v.ErrMsg = "invalid", "code required for new book"
					validatedChan <- v
					continue
				}

				if _, exists := existingByCode[row.code]; exists {
					v.Status, v.ErrMsg = "invalid", fmt.Sprintf("code '%s' already exists", row.code)
					validatedChan <- v
					continue
				}

				v.Status = "create"
				validatedChan <- v
			}
		}
	})

	// Stage 3: Collector
	var validatedRows []ValidatedBookRow

	g.Go(func() error {
		for v := range validatedChan {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			validatedRows = append(validatedRows, v)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return validatedRows, nil
}

func (u Usecase) PreviewImportBooks(ctx context.Context, libID uuid.UUID, r io.Reader, filename string) (PreviewImportBooksResult, error) {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return PreviewImportBooksResult{}, fmt.Errorf("user id not found in context")
	}

	_, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
		UserID:     userID.String(),
		LibraryIDs: uuid.UUIDs{libID},
	})
	if err != nil {
		return PreviewImportBooksResult{}, err
	}

	g, ctx := errgroup.WithContext(ctx)

	var (
		res PreviewImportBooksResult
		mu  sync.Mutex
	)

	data, err := io.ReadAll(r)
	if err != nil {
		return res, err
	}

	g.Go(func() error {
		mu.Lock()
		res.Path = "/private/" + libID.String() + "/imports/" + filename
		mu.Unlock()
		return u.fileStorageProvider.UploadFile(ctx, res.Path, data)
	})

	g.Go(func() error {
		validatedRows, err := u.validateImportBooksCSV(ctx, libID, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		var rows []PreviewImportBooksRow
		var createCount, updateCount, invalidCount int

		for _, v := range validatedRows {
			var idStr, errStr *string
			if v.ID != nil {
				s := v.ID.String()
				idStr = &s
			}
			if v.ErrMsg != "" {
				errStr = &v.ErrMsg
			}

			rows = append(rows, PreviewImportBooksRow{
				ID:     idStr,
				Code:   v.Code,
				Title:  v.Title,
				Author: v.Author,
				Status: v.Status,
				Error:  errStr,
			})

			switch v.Status {
			case "create":
				createCount++
			case "update":
				updateCount++
			case "invalid":
				invalidCount++
			}
		}

		mu.Lock()
		res.Summary = PreviewImportBooksSummary{
			CreatedCount: createCount,
			UpdatedCount: updateCount,
			InvalidCount: invalidCount,
		}
		res.Rows = rows
		mu.Unlock()

		return nil
	})

	if err := g.Wait(); err != nil {
		return res, err
	}

	return res, nil
}

func (u Usecase) ConfirmImportBooks(ctx context.Context, libID uuid.UUID, path string) (string, error) {

	_, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return "", fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return "", fmt.Errorf("user id not found in context")
	}
	staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
		UserID:     userID.String(),
		LibraryIDs: uuid.UUIDs{libID},
		Limit:      1,
	})
	if err != nil {
		return "", err
	}
	if len(staffs) == 0 {
		return "", fmt.Errorf("user %s not staff of library %s", userID, libID)
	}

	b, err := json.Marshal(map[string]string{
		"path":       path,
		"library_id": libID.String(),
	})
	if err != nil {
		return "", err
	}

	job, err := u.CreateJob(ctx, Job{
		Type:    "import:books",
		StaffID: staffs[0].ID,
		Status:  "PENDING",
		Payload: b,
	})
	if err != nil {
		return "", err
	}
	return job.ID.String(), nil

}

func (u Usecase) ProcessImportBooksJob(ctx context.Context, jobID uuid.UUID) error {

	// 1. Get job from database
	job, err := u.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// 2. Parse job payload
	var payload struct {
		Path  string    `json:"path"`
		LibID uuid.UUID `json:"library_id"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse job payload: %w", err)
	}

	// 3. Update job status to PROCESSING
	now := time.Now()
	job.Status = "PROCESSING"
	job.StartedAt = &now
	if _, err := u.repo.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job to PROCESSING: %w", err)
	}

	// 4. Execute the export work
	res, err := u.executeImportBooks(ctx, payload.LibID, payload.Path)
	if err != nil {
		// Update job status to FAILED
		finished := time.Now()
		job.Status = "FAILED"
		job.Error = err.Error()
		job.FinishedAt = &finished
		u.repo.UpdateJob(ctx, job)
		return fmt.Errorf("export failed: %w", err)
	}

	// 5. Update job status to COMPLETED
	finished := time.Now()
	job.Status = "COMPLETED"
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal job result: %w", err)
	}
	job.Result = data
	job.FinishedAt = &finished
	if _, err := u.repo.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job to COMPLETED: %w", err)
	}

	// 6. Send notification to staff
	go func() {
		if job.Staff != nil {
			if err := u.CreateNotification(context.Background(), Notification{
				UserID:        job.Staff.UserID,
				Title:         "Import Completed",
				Message:       "Your book import job has completed successfully.",
				ReferenceType: "IMPORT_BOOKS",
				ReferenceID:   &job.ID,
			}); err != nil {
				fmt.Printf("failed to send notification for job %s: %v\n", job.ID, err)
			}
		}
	}()

	return nil
}

type ImportBooksResult struct {
	TotalRows    int               `json:"total_rows"`
	SuccessCount int               `json:"success_count"`
	FailedCount  int               `json:"failed_count"`
	SkippedCount int               `json:"skipped_count"`
	CreatedBooks []uuid.UUID       `json:"created_books"`
	UpdatedBooks []uuid.UUID       `json:"updated_books"`
	FailedRows   []ImportFailedRow `json:"failed_rows"`
}

type ImportFailedRow struct {
	RowNum int    `json:"row_num"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Error  string `json:"error"`
}

func (u Usecase) executeImportBooks(ctx context.Context, libID uuid.UUID, path string) (ImportBooksResult, error) {

	r, err := u.fileStorageProvider.GetReader(ctx, path)
	if err != nil {
		return ImportBooksResult{}, fmt.Errorf("failed to get file reader: %w", err)
	}
	defer r.Close()

	// Validate the CSV and get validated rows
	validatedRows, err := u.validateImportBooksCSV(ctx, libID, r)
	if err != nil {
		return ImportBooksResult{}, err
	}

	result := ImportBooksResult{
		TotalRows:    len(validatedRows),
		CreatedBooks: []uuid.UUID{},
		UpdatedBooks: []uuid.UUID{},
		FailedRows:   []ImportFailedRow{},
	}

	// Process each validated row
	for _, v := range validatedRows {
		// Skip invalid rows
		if v.Status == "invalid" {
			result.SkippedCount++
			result.FailedRows = append(result.FailedRows, ImportFailedRow{
				RowNum: v.RowNum,
				Code:   v.Code,
				Title:  v.Title,
				Error:  v.ErrMsg,
			})
			continue
		}

		// Handle create
		if v.Status == "create" {
			book, err := u.repo.CreateBook(ctx, Book{
				Title:     v.Title,
				Author:    v.Author,
				Year:      v.Year,
				Code:      v.Code,
				LibraryID: libID,
			})
			if err != nil {
				result.FailedCount++
				result.FailedRows = append(result.FailedRows, ImportFailedRow{
					RowNum: v.RowNum,
					Code:   v.Code,
					Title:  v.Title,
					Error:  fmt.Sprintf("failed to create: %v", err),
				})
				continue
			}
			result.SuccessCount++
			result.CreatedBooks = append(result.CreatedBooks, book.ID)
		}

		// Handle update
		if v.Status == "update" && v.ID != nil {
			book, err := u.repo.UpdateBook(ctx, *v.ID, Book{
				Title:  v.Title,
				Author: v.Author,
				Year:   v.Year,
				Code:   v.Code,
			})
			if err != nil {
				result.FailedCount++
				result.FailedRows = append(result.FailedRows, ImportFailedRow{
					RowNum: v.RowNum,
					Code:   v.Code,
					Title:  v.Title,
					Error:  fmt.Sprintf("failed to update: %v", err),
				})
				continue
			}
			result.SuccessCount++
			result.UpdatedBooks = append(result.UpdatedBooks, book.ID)
		}
	}

	return result, nil
}
