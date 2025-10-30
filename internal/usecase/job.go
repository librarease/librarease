package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type Job struct {
	ID         uuid.UUID
	Type       string
	StaffID    uuid.UUID
	Status     string
	Payload    []byte
	Result     []byte
	Error      string
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time

	Staff *Staff
}

type ListJobsOption struct {
	Skip   int
	Limit  int
	SortBy string
	SortIn string

	Types    []string
	StaffIDs uuid.UUIDs
	Statuses []string

	LibraryID uuid.UUID
}

// ListJobs retrieves a list of jobs with authorization checks based on user role
func (u Usecase) ListJobs(ctx context.Context, opt ListJobsOption) ([]Job, int, error) {
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
		fmt.Println("[DEBUG] global user - staff access only")
		// Verify user is staff of the specified library
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{opt.LibraryID},
			Limit:      1,
		})
		if err != nil {
			return nil, 0, err
		}

		// User must be staff of the specified library
		if len(staffs) == 0 {
			fmt.Println("[DEBUG] user is not staff of the specified library")
			return nil, 0, fmt.Errorf("unauthorized: not a staff member of this library")
		}

		fmt.Println("[DEBUG] staff verified - access to all jobs in library")
	}

	return u.repo.ListJobs(ctx, opt)
}

// GetJobByID retrieves a single job by ID with authorization checks
func (u Usecase) GetJobByID(ctx context.Context, id uuid.UUID) (Job, error) {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Job{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Job{}, fmt.Errorf("user id not found in context")
	}

	job, err := u.repo.GetJobByID(ctx, id)
	if err != nil {
		return Job{}, err
	}

	switch role {
	case "SUPERADMIN", "ADMIN":
		// ALLOW ALL
		return job, nil
	case "USER":
		// Check if user is staff of the library that owns this job
		// First need to get the staff who created the job to know which library
		if job.Staff == nil {
			return Job{}, fmt.Errorf("job staff information not loaded")
		}

		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{job.Staff.LibraryID},
			Limit:      1,
		})
		if err != nil {
			return Job{}, err
		}

		if len(staffs) == 0 {
			fmt.Println("[DEBUG] user is not staff of the job's library")
			return Job{}, fmt.Errorf("unauthorized: not a staff member of this library")
		}

		fmt.Println("[DEBUG] staff verified - access granted to job")
		return job, nil
	}

	return job, nil
}

// CreateJob creates a new job and enqueues it to the async queue
func (u Usecase) CreateJob(ctx context.Context, job Job) (Job, error) {
	// Set default status if not provided
	if job.Status == "" {
		job.Status = "PENDING"
	}

	// Create job record in database
	createdJob, err := u.repo.CreateJob(ctx, job)
	if err != nil {
		return Job{}, err
	}

	// Enqueue the job task to the async queue
	if err := u.queueClient.EnqueueJob(ctx, createdJob.ID, createdJob.Type, createdJob.Payload); err != nil {
		// Log error but don't fail the job creation
		// The job is in PENDING state and can be retried manually or by a cleanup job
		fmt.Printf("[Job] Failed to enqueue job %s: %v\n", createdJob.ID, err)
	}
	fmt.Printf("[Job] Successfully enqueued job %s (type: %s)\n", createdJob.ID, createdJob.Type)

	return createdJob, nil
}

// UpdateJob updates an existing job
func (u Usecase) UpdateJob(ctx context.Context, job Job) (Job, error) {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Job{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Job{}, fmt.Errorf("user id not found in context")
	}

	// Get existing job for authorization check
	existingJob, err := u.repo.GetJobByID(ctx, job.ID)
	if err != nil {
		return Job{}, err
	}

	switch role {
	case "SUPERADMIN", "ADMIN":
		// ALLOW ALL
	case "USER":
		// Verify user is the staff who created the job
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			Limit:  500,
		})
		if err != nil {
			return Job{}, err
		}

		hasAccess := false
		for _, staff := range staffs {
			if staff.ID == existingJob.StaffID {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return Job{}, fmt.Errorf("unauthorized: cannot update this job")
		}
	}

	return u.repo.UpdateJob(ctx, job)
}

// DeleteJob soft deletes a job
func (u Usecase) DeleteJob(ctx context.Context, id uuid.UUID) error {
	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}

	// Get existing job for authorization check
	existingJob, err := u.repo.GetJobByID(ctx, id)
	if err != nil {
		return err
	}

	switch role {
	case "SUPERADMIN", "ADMIN":
		// ALLOW ALL
	case "USER":
		// Verify user is the staff who created the job
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			Limit:  500,
		})
		if err != nil {
			return err
		}

		hasAccess := false
		for _, staff := range staffs {
			if staff.ID == existingJob.StaffID {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return fmt.Errorf("unauthorized: cannot delete this job")
		}
	}

	return u.repo.DeleteJob(ctx, id)
}

func (u Usecase) DownloadJobAsset(ctx context.Context, id uuid.UUID) (string, error) {

	job, err := u.repo.GetJobByID(ctx, id)
	if err != nil {
		return "", err
	}

	if job.Status != "COMPLETED" {
		return "", fmt.Errorf("job is not completed")
	}

	var res struct {
		Path string `json:"path"`
	}

	var b []byte

	switch job.Type {
	case "export:borrowings":
		b = job.Result
	case "import:books":
		b = job.Payload
	default:
		return "", fmt.Errorf("unsupported job type for download: %s", job.Type)
	}

	if err := json.Unmarshal(b, &res); err != nil {
		return "", fmt.Errorf("failed to parse job asset: %w", err)
	}

	return u.fileStorageProvider.GetPresignedURL(ctx, res.Path)
}
