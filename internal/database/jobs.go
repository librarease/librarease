package database

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Job struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Type       string          `gorm:"column:type;type:varchar(255);NOT NULL"`
	StaffID    uuid.UUID       `gorm:"column:staff_id;type:uuid"`
	Status     string          `gorm:"column:status;type:varchar(255);NOT NULL"`
	Payload    datatypes.JSON  `gorm:"column:payload"`
	Result     datatypes.JSON  `gorm:"column:result"`
	Error      string          `gorm:"column:error;type:text"`
	StartedAt  *time.Time      `gorm:"column:started_at"`
	FinishedAt *time.Time      `gorm:"column:finished_at"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at"`

	Staff *Staff `gorm:"foreignKey:StaffID;references:ID"`
}

func (Job) TableName() string {
	return "jobs"
}

func (s *service) CreateJob(ctx context.Context, job usecase.Job) (usecase.Job, error) {
	j := Job{
		Type:    job.Type,
		StaffID: job.StaffID,
		Status:  job.Status,
		Payload: job.Payload,
	}
	if err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&j).Error; err != nil {
		return usecase.Job{}, err
	}

	return j.ConvertToUsecase(), nil
}

func (s *service) ListJobs(ctx context.Context, opt usecase.ListJobsOption) ([]usecase.Job, int, error) {
	var (
		jobs  []Job
		ujobs []usecase.Job
		count int64
	)

	db := s.db.Model([]Job{}).WithContext(ctx)

	if opt.LibraryID != uuid.Nil {
		db = db.Joins("JOIN staffs ON jobs.staff_id = staffs.id").
			Where("staffs.library_id = ?", opt.LibraryID)
	}
	if opt.Types != nil {
		db = db.Where("type IN ?", opt.Types)
	}
	if opt.StaffIDs != nil {
		db = db.Where("staff_id IN ?", opt.StaffIDs)
	}
	if opt.Statuses != nil {
		db = db.Where("status IN ?", opt.Statuses)
	}

	var (
		orderIn = "DESC"
		orderBy = "created_at"
	)

	if slices.Contains([]string{"ASC", "DESC"}, opt.SortIn) {
		orderIn = opt.SortIn
	}
	if slices.Contains([]string{"created_at", "updated_at", "started_at", "finished_at"}, opt.SortBy) {
		orderBy = opt.SortBy
	}

	db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: orderBy}, Desc: orderIn == "DESC"})

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if err := db.Preload("Staff").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	for _, job := range jobs {
		uj := job.ConvertToUsecase()
		if job.Staff != nil {
			staff := job.Staff.ConvertToUsecase()
			uj.Staff = &staff
		}
		ujobs = append(ujobs, uj)
	}

	return ujobs, int(count), nil
}

func (s *service) UpdateJob(ctx context.Context, job usecase.Job) (usecase.Job, error) {
	if err := s.db.
		WithContext(ctx).
		Clauses(clause.Returning{}).
		Model(&Job{}).
		Where("id = ?", job.ID).
		Updates(Job{
			Type:       job.Type,
			StaffID:    job.StaffID,
			Status:     job.Status,
			Payload:    job.Payload,
			Result:     job.Result,
			Error:      job.Error,
			StartedAt:  job.StartedAt,
			FinishedAt: job.FinishedAt,
		}).Error; err != nil {

		return usecase.Job{}, err
	}

	return job, nil
}

func (s *service) GetJobByID(ctx context.Context, id uuid.UUID) (usecase.Job, error) {
	var job Job
	if err := s.db.
		WithContext(ctx).
		Preload("Staff").
		First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return usecase.Job{}, usecase.ErrNotFound{
				ID:      id,
				Code:    "job_not_found",
				Message: "job " + id.String() + " not found",
			}
		}
		return usecase.Job{}, err
	}

	uj := job.ConvertToUsecase()
	if job.Staff != nil {
		staff := job.Staff.ConvertToUsecase()
		uj.Staff = &staff
	}

	return uj, nil
}

func (s *service) DeleteJob(ctx context.Context, id uuid.UUID) error {
	return s.db.
		WithContext(ctx).
		Delete(&Job{}, "id = ?", id).Error
}

// Convert core model to usecase model
func (j Job) ConvertToUsecase() usecase.Job {
	var d *time.Time
	if j.DeletedAt != nil {
		d = &j.DeletedAt.Time
	}
	return usecase.Job{
		ID:         j.ID,
		Type:       j.Type,
		StaffID:    j.StaffID,
		Status:     j.Status,
		Payload:    j.Payload,
		Result:     j.Result,
		Error:      j.Error,
		StartedAt:  j.StartedAt,
		FinishedAt: j.FinishedAt,
		CreatedAt:  j.CreatedAt,
		UpdatedAt:  j.UpdatedAt,
		DeletedAt:  d,
	}
}
