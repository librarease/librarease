package usecase

import (
	"context"
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

// CreateJob creates a new job and returns it
// TODO: Integrate with async queue - enqueue task after job creation
// Example: After successfully creating the job, send it to a message queue
// like Redis, RabbitMQ, or AWS SQS for async processing by worker processes
func (u Usecase) CreateJob(ctx context.Context, job Job) (Job, error) {
	// role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	// if !ok {
	// 	return Job{}, fmt.Errorf("user role not found in context")
	// }
	// userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	// if !ok {
	// 	return Job{}, fmt.Errorf("user id not found in context")
	// }

	// // Verify the staff creating the job exists and belongs to the user
	// switch role {
	// case "SUPERADMIN", "ADMIN":
	// 	// ALLOW ALL - admins can create jobs for any staff
	// case "USER":
	// 	// Verify user is staff
	// 	staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
	// 		UserID: userID.String(),
	// 		Limit:  500,
	// 	})
	// 	if err != nil {
	// 		return Job{}, err
	// 	}

	// 	// Check if the job's StaffID matches one of user's staff records
	// 	hasAccess := false
	// 	for _, staff := range staffs {
	// 		if staff.ID == job.StaffID {
	// 			hasAccess = true
	// 			break
	// 		}
	// 	}

	// 	if !hasAccess {
	// 		return Job{}, fmt.Errorf("unauthorized: cannot create job for this staff")
	// 	}
	// }

	// Set default status if not provided
	if job.Status == "" {
		job.Status = "PENDING"
	}

	createdJob, err := u.repo.CreateJob(ctx, job)
	if err != nil {
		return Job{}, err
	}

	// TODO: Async Queue Integration
	// After successful job creation, enqueue the task to an async queue
	// Example pseudo-code:
	//
	// if err := u.queue.Enqueue(ctx, QueueMessage{
	//     JobID:   createdJob.ID,
	//     Type:    createdJob.Type,
	//     Payload: createdJob.Payload,
	// }); err != nil {
	//     // Log error but don't fail the job creation
	//     fmt.Printf("failed to enqueue job %s: %v\n", createdJob.ID, err)
	// }
	//
	// Worker process would:
	// 1. Dequeue the message
	// 2. Update job status to "PROCESSING" via UpdateJob
	// 3. Execute the task
	// 4. Update job with result/error and status "COMPLETED"/"FAILED"

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
