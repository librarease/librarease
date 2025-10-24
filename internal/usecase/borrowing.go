package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/librarease/librarease/internal/config"

	"github.com/google/uuid"
)

type Borrowing struct {
	ID             uuid.UUID
	BookID         uuid.UUID
	SubscriptionID uuid.UUID
	StaffID        uuid.UUID
	BorrowedAt     time.Time
	DueAt          time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	Book         *Book
	Subscription *Subscription
	Staff        *Staff
	Returning    *Returning
	Lost         *Lost
}

// BorrowingSummary contains minimal borrowing info for efficient notification processing
type BorrowingSummary struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DueAt       time.Time
	BookTitle   string
	LibraryName string
}

// NotificationFiltersOption contains precise filters for notification processing
type NotificationFiltersOption struct {
	DueAtFrom  *time.Time
	DueAtTo    *time.Time
	LibraryIDs []uuid.UUID
	Limit      int
}

type ListBorrowingsOption struct {
	Skip   int
	Limit  int
	SortBy string
	SortIn string

	BookIDs         uuid.UUIDs
	SubscriptionIDs uuid.UUIDs
	BorrowStaffIDs  uuid.UUIDs
	ReturnStaffIDs  uuid.UUIDs
	MembershipIDs   uuid.UUIDs
	LibraryIDs      uuid.UUIDs
	UserIDs         uuid.UUIDs
	ReturningIDs    uuid.UUIDs
	BorrowedAt      time.Time
	DueAt           time.Time
	ReturnedAt      *time.Time
	LostAt          *time.Time
	IsActive        bool
	IsOverdue       bool
	IsReturned      bool
	IsLost          bool
	BorrowedAtFrom  *time.Time
	BorrowedAtTo    *time.Time
}

// TODO: separate client and admin borrowing list route
func (u Usecase) ListBorrowings(ctx context.Context, opt ListBorrowingsOption) ([]Borrowing, int, error) {

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return nil, 0, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return nil, 0, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		fmt.Println("[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		fmt.Println("[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		fmt.Println("[DEBUG] global user")
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			// Using a limit of 500 for now, adjust as needed based on expected data size
			Limit: 500,
		})
		if err != nil {
			return nil, 0, err
		}
		// user is not staff
		if len(staffs) == 0 {
			fmt.Println("[DEBUG] user is not staff, filtering by user id")
			opt.UserIDs = uuid.UUIDs{userID}
			break
		}

		// user is staff
		fmt.Println("[DEBUG] user is staff")
		var staffLibIDs uuid.UUIDs
		for _, staff := range staffs {
			staffLibIDs = append(staffLibIDs, staff.LibraryID)
		}
		// user is staff, filtering by library ids
		if len(opt.LibraryIDs) == 0 {
			// user is staff, filters default to assigned libraries
			fmt.Println("[DEBUG] filtering by default assigned libraries")
			opt.LibraryIDs = staffLibIDs
			break
		}

		fmt.Println("[DEBUG] filtering by library ids query")
		var intersectLibIDs uuid.UUIDs
		for _, id := range opt.LibraryIDs {
			// filter out library ids that are not assigned to the staff
			if slices.Contains(staffLibIDs, id) {
				intersectLibIDs = append(intersectLibIDs, id)
			}
		}

		if len(intersectLibIDs) == 0 {
			// user is filtering by library ids but none of the ids are assigned to the staff
			fmt.Println("[DEBUG] staff filters by lib ids but none assigned")
			opt.LibraryIDs = staffLibIDs
			break
		}

		// user is filtering by library ids and some of the ids are assigned to the staff
		fmt.Println("[DEBUG] staff filters by lib ids and some assigned")
		opt.LibraryIDs = intersectLibIDs
	}

	borrows, total, err := u.repo.ListBorrowings(ctx, opt)
	if err != nil {
		return nil, 0, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	for i, borrow := range borrows {
		if b := borrow.Book; b != nil && b.Cover != "" {
			borrows[i].Book.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, b.ID, b.Cover)
		}
	}
	return borrows, total, nil
}

type ErrNotFound struct {
	ID      uuid.UUID
	Code    string
	Message string
}

func (e ErrNotFound) Error() string {
	return e.Message
}

func (u Usecase) GetBorrowingByID(ctx context.Context, id uuid.UUID) (Borrowing, error) {
	borrow, err := u.repo.GetBorrowingByID(ctx, id)
	if err != nil {
		return Borrowing{}, err
	}

	publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)
	if b := borrow.Book; b != nil && b.Cover != "" {
		borrow.Book.Cover = fmt.Sprintf("%s/books/%s/cover/%s", publicURL, b.ID, b.Cover)
	}

	if borrow.Subscription != nil &&
		borrow.Subscription.Membership != nil &&
		borrow.Subscription.Membership.Library != nil &&
		borrow.Subscription.Membership.Library.Logo != "" {
		borrow.Subscription.Membership.Library.Logo = fmt.Sprintf("%s/libraries/%s/logo/%s",
			publicURL,
			borrow.Subscription.Membership.Library.ID,
			borrow.Subscription.Membership.Library.Logo,
		)
	}

	return borrow, nil
}

func (u Usecase) CreateBorrowing(ctx context.Context, borrow Borrowing) (Borrowing, error) {

	// 1. Check if the membership subscription is still active
	s, err := u.repo.GetSubscriptionByID(ctx, borrow.SubscriptionID)
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrMembershipExpired
	if s.ExpiresAt.Before(time.Now()) {
		fmt.Println(s.ExpiresAt, time.Now())
		return Borrowing{}, fmt.Errorf("membership subscription %s expired", s.ID)
	}

	if s.UsageLimit > 0 {
		_, usageCount, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
			SubscriptionIDs: uuid.UUIDs{s.ID},
		})
		if err != nil {
			return Borrowing{}, err
		}
		if usageCount >= s.UsageLimit {
			return Borrowing{}, fmt.Errorf("subscription %s has reached the usage limit %d", s.ID, s.UsageLimit)
		}
	}

	// 2. Check if the user has reached the maximum borrowing limit
	_, activeBorrowCount, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		SubscriptionIDs: uuid.UUIDs{s.ID},
		IsActive:        true,
	})
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrActiveLoanLimitReached
	if s.ActiveLoanLimit <= activeBorrowCount {
		return Borrowing{}, fmt.Errorf("user %s has reached the active loan limit", s.UserID)
	}

	// 3. Check if the book is from the library
	book, err := u.repo.GetBookByID(ctx, borrow.BookID)
	if err != nil {
		return Borrowing{}, err
	}
	m, err := u.repo.GetMembershipByID(ctx, s.MembershipID)
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrBookNotInLibrary
	if book.LibraryID != m.LibraryID {
		return Borrowing{}, fmt.Errorf("book %s is not in library %s", book.ID, m.LibraryID)
	}

	// 4. Check if the book is available
	_, activeBookCount, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BookIDs:  uuid.UUIDs{borrow.BookID},
		IsActive: true,
	})
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrBookNotAvailable
	if activeBookCount > 0 {
		return Borrowing{}, fmt.Errorf("book %s is not available", borrow.BookID)
	}

	// 4. Check if the book is available
	_, lostBookCount, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		BookIDs: uuid.UUIDs{borrow.BookID},
		IsLost:  true,
	})
	if err != nil {
		return Borrowing{}, err
	}
	// TODO: ErrBookNotAvailable
	if lostBookCount > 0 {
		return Borrowing{}, fmt.Errorf("book %s is not available (lost)", borrow.BookID)
	}

	// 5. Check if staff exists
	staff, err := u.repo.GetStaffByID(ctx, borrow.StaffID)
	if err != nil {
		return Borrowing{}, err
	}
	if staff.LibraryID != m.LibraryID {
		return Borrowing{}, fmt.Errorf("staff %s is not from library %s", staff.ID, m.LibraryID)
	}

	// 6. All checks passed, create borrowing
	// Set the borrowed at time if not set
	if borrow.BorrowedAt.IsZero() {
		borrow.BorrowedAt = time.Now()
	}
	// Set the due at time if not set
	if borrow.DueAt.IsZero() {
		borrow.DueAt = time.Now().AddDate(0, 0, s.LoanPeriod)
	}

	bw, err := u.repo.CreateBorrowing(ctx, borrow)
	if err != nil {
		return Borrowing{}, err
	}

	go func() {
		// if err := u.SendBorrowingEmail(context.Background(), bw.ID); err != nil {
		// 	fmt.Printf("borrowing: failed to send email: %v\n", err)
		// }

		if err := u.CreateNotification(context.Background(), Notification{
			Title: "Book Borrowed",
			Message: fmt.Sprintf("You have successfully borrowed %s from %s. Please return it by %s. Happy reading!",
				book.Title,
				book.Library.Name,
				bw.DueAt.Format("2006-01-02 03:04 PM")),
			UserID:        s.UserID,
			ReferenceType: "BORROWING",
			ReferenceID:   &bw.ID,
		}); err != nil {
			fmt.Printf("borrowing: failed to create notification: %v\n", err)
		}

	}()

	return bw, nil
}

func (u Usecase) UpdateBorrowing(ctx context.Context, borrow Borrowing) (Borrowing, error) {
	return u.repo.UpdateBorrowing(ctx, borrow)
}

func (u Usecase) DeleteBorrowing(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteBorrowing(ctx, id)
}

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

// ProcessExportBorrowingsJob executes the actual export work for a job
// This method is called by the queue worker to process the export asynchronously
func (u Usecase) ProcessExportBorrowingsJob(ctx context.Context, jobID uuid.UUID) error {
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

	// 4. Execute the export work
	res, err := u.executeExport(ctx, payload)
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
	job.Result = res
	job.FinishedAt = &finished
	if _, err := u.repo.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job to COMPLETED: %w", err)
	}

	// 6. Send notification to staff
	go func() {
		if job.Staff != nil {
			if err := u.CreateNotification(context.Background(), Notification{
				UserID:        job.Staff.UserID,
				Title:         "Export Ready",
				Message:       "Your borrowings export is ready for download",
				ReferenceType: "EXPORT_BORROWING",
				ReferenceID:   &job.ID,
			}); err != nil {
				fmt.Printf("failed to send notification for job %s: %v\n", job.ID, err)
			}
		}
	}()

	return nil
}

// executeExport performs the actual export logic
func (u Usecase) executeExport(ctx context.Context, payload ExportBorrowingsJobPayload) ([]byte, error) {

	// 1. Query borrowings with filters
	borrowings, _, err := u.repo.ListBorrowings(ctx, ListBorrowingsOption{
		LibraryIDs:     uuid.UUIDs{payload.LibraryID},
		IsActive:       payload.IsActive,
		IsOverdue:      payload.IsOverdue,
		IsReturned:     payload.IsReturned,
		IsLost:         payload.IsLost,
		BorrowedAtFrom: payload.BorrowedAtFrom,
		BorrowedAtTo:   payload.BorrowedAtTo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list borrowings: %w", err)
	}

	// 2. Generate CSV file
	csvData := generateCSV(borrowings)

	// 3. Upload to file storage
	fileName := fmt.Sprintf("borrowings-export-%s-%s.csv",
		payload.LibraryID, time.Now().Format("20060102-150405"))
	path := "/private/" + payload.LibraryID.String() + "/exports/" + fileName

	if err := u.fileStorageProvider.UploadFile(ctx, path, csvData); err != nil {
		return nil, fmt.Errorf("failed to upload export file: %w", err)
	}

	return json.Marshal(map[string]any{
		"path": path,
		"name": fileName,
		"size": len(csvData),
	})
}

func generateCSV(borrowings []Borrowing) []byte {
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
