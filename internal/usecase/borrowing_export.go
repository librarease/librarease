package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ExportBorrowingsOption struct {
	LibraryID uuid.UUID

	IsActive       bool
	IsOverdue      bool
	IsReturned     bool
	IsLost         bool
	BorrowedAtFrom *time.Time
	BorrowedAtTo   *time.Time
}
type ExportBorrowingsJobPayload struct {
	LibraryID      uuid.UUID  `json:"library_id"`
	IsActive       bool       `json:"is_active"`
	IsOverdue      bool       `json:"is_overdue"`
	IsReturned     bool       `json:"is_returned"`
	IsLost         bool       `json:"is_lost"`
	BorrowedAtFrom *time.Time `json:"borrowed_at_from,omitempty"`
	BorrowedAtTo   *time.Time `json:"borrowed_at_to,omitempty"`
}

func (u Usecase) ExportBorrowings(ctx context.Context, opt ExportBorrowingsOption) (string, error) {
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
		LibraryIDs: uuid.UUIDs{opt.LibraryID},
		Limit:      1,
	})
	if err != nil {
		return "", err
	}
	if len(staffs) == 0 {
		return "", fmt.Errorf("user %s not staff of library %s", userID, opt.LibraryID)
	}
	b, err := json.Marshal(ExportBorrowingsJobPayload(opt))
	if err != nil {
		return "", err
	}
	job, err := u.CreateJob(ctx, Job{
		Type:    "export:borrowings",
		StaffID: staffs[0].ID,
		Status:  "PENDING",
		Payload: b,
	})
	if err != nil {
		return "", err
	}
	return job.ID.String(), nil
}

func (u Usecase) ProcessExportBorrowingsJob(ctx context.Context, jobID uuid.UUID) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "usecase.job.process_export_borrowings",
		attribute.String("job.id", jobID.String()),
		attribute.String("job.type", "export:borrowings"),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	// 1. Get job from database
	job, err := u.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// 2. Parse job payload
	var payload ExportBorrowingsJobPayload
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
	span.AddEvent("job.status_changed", trace.WithAttributes(attribute.String("job.status", "PROCESSING")))

	// 4. Execute the export work
	res, err := u.executeExport(ctx, payload)
	if err != nil {
		// Update job status to FAILED
		finished := time.Now()
		job.Status = "FAILED"
		job.Error = err.Error()
		job.FinishedAt = &finished
		if _, updateErr := u.repo.UpdateJob(ctx, job); updateErr != nil {
			span.AddEvent("job.failure_status_update_failed", trace.WithAttributes(attribute.String("err", updateErr.Error())))
			u.logger.ErrorContext(ctx, "failed to update export job to failed",
				slog.String("job_id", job.ID.String()),
				slog.String("err", updateErr.Error()))
		}
		span.AddEvent("job.status_changed", trace.WithAttributes(attribute.String("job.status", "FAILED")))
		return fmt.Errorf("export failed: %w", err)
	}

	// 5. Update job status to COMPLETED
	finished := time.Now()
	job.Status = "COMPLETED"
	job.Result = res
	job.FinishedAt = &finished
	if _, err := u.repo.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job to COMPLETED: %w", err)
	}
	span.AddEvent("job.status_changed", trace.WithAttributes(attribute.String("job.status", "COMPLETED")))

	// 6. Send notification to staff
	if job.Staff != nil {
		if err := u.EnqueueNotification(ctx, Notification{
			UserID:        job.Staff.UserID,
			Title:         "Export Ready",
			Message:       "Your borrowings export is ready for download",
			ReferenceType: "EXPORT_BORROWING",
			ReferenceID:   &job.ID,
		}); err != nil {
			span.AddEvent("job.notification_enqueue_failed", trace.WithAttributes(attribute.String("err", err.Error())))
			u.logger.ErrorContext(ctx, "failed to enqueue export job notification",
				slog.String("job_id", job.ID.String()),
				slog.String("err", err.Error()))
		}
	}

	return nil
}

func (u Usecase) executeExport(ctx context.Context, payload ExportBorrowingsJobPayload) (result []byte, err error) {
	ctx, span := telemetry.StartSpan(ctx, "usecase.borrowings.execute_export",
		attribute.String("library.id", payload.LibraryID.String()),
		attribute.Bool("borrowings.filter.is_active", payload.IsActive),
		attribute.Bool("borrowings.filter.is_overdue", payload.IsOverdue),
		attribute.Bool("borrowings.filter.is_returned", payload.IsReturned),
		attribute.Bool("borrowings.filter.is_lost", payload.IsLost),
	)
	defer func() { telemetry.EndSpan(span, err) }()

	// 1. Query borrowings with filters
	borrowings, _, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BorrowingsOption: BorrowingsOption{
			LibraryIDs:     uuid.UUIDs{payload.LibraryID},
			IsActive:       payload.IsActive,
			IsOverdue:      payload.IsOverdue,
			IsReturned:     payload.IsReturned,
			IsLost:         payload.IsLost,
			BorrowedAtFrom: payload.BorrowedAtFrom,
			BorrowedAtTo:   payload.BorrowedAtTo,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list borrowings: %w", err)
	}
	span.SetAttributes(attribute.Int("borrowings.export.row_count", len(borrowings)))

	// 2. Generate CSV file
	csvData := generateBorrowingCSV(borrowings)

	// 3. Upload to file storage
	fileName := fmt.Sprintf("borrowings-export-%s.csv", time.Now().Format("20060102-150405"))
	path := payload.LibraryID.String() + "/exports/" + fileName

	if err := u.fileStorageProvider.UploadFile(ctx, path, csvData); err != nil {
		return nil, fmt.Errorf("failed to upload export file: %w", err)
	}
	span.SetAttributes(
		attribute.String("borrowings.export.path", path),
		attribute.Int("borrowings.export.size_bytes", len(csvData)),
	)

	return json.Marshal(map[string]any{
		"path": path,
		"name": fileName,
		"size": len(csvData),
	})
}

func generateBorrowingCSV(borrowings []Borrowing) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	// Write header
	writer.Write([]string{"User", "Book", "Status", "Borrowed At", "Due At", "Returned At", "Lost At"})

	var user, book, status, returnedAt, lostAt string
	// Write rows
	for _, b := range borrowings {
		switch {
		case b.Lost != nil:
			status = "Lost"
			lostAt = b.Lost.ReportedAt.UTC().Format("2006-01-02 15:04")

		case b.Returning != nil:
			status = "Returned"
			returnedAt = b.Returning.ReturnedAt.UTC().Format("2006-01-02 15:04")

		case time.Now().After(b.DueAt):
			status = "Overdue"
		default:
			status = "Active"
		}

		if b.Subscription != nil && b.Subscription.User != nil {
			user = b.Subscription.User.Name
		}
		if b.Book != nil {
			book = b.Book.Title
		}
		writer.Write([]string{
			user,
			book,
			status,
			b.BorrowedAt.UTC().Format("2006-01-02 15:04"),
			b.DueAt.UTC().Format("2006-01-02 15:04"),
			returnedAt,
			lostAt,
		})
		// Reset for next row
		user, book, status, returnedAt, lostAt = "", "", "", "", ""
	}
	writer.Flush()
	return buf.Bytes()
}
