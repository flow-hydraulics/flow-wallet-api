package jobs

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db}
}

func (s *GormStore) Jobs(o datastore.ListOptions) (jj []Job, err error) {
	err = s.db.
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&jj).Error
	return
}

func (s *GormStore) Job(id uuid.UUID) (j Job, err error) {
	err = s.db.First(&j, "id = ?", id).Error
	return
}

func (s *GormStore) InsertJob(j *Job) error {
	return s.db.Create(j).Error
}

func (s *GormStore) UpdateJob(j *Job) error {
	return s.db.Save(j).Error
}

func (s *GormStore) IncreaseExecCount(j *Job) error {
	return s.db.Model(j).
		Where("id = ? AND exec_count = ? AND updated_at = ?", j.ID, j.ExecCount, j.UpdatedAt).
		Update("exec_count", j.ExecCount+1).Error
}

func (s *GormStore) SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod time.Duration, o datastore.ListOptions) (jj []Job, err error) {
	t0 := time.Now()
	tAccepted := t0.Add(-1 * acceptedGracePeriod)
	tReschedulable := t0.Add(-1 * reSchedulableGracePeriod)

	err = s.db.
		Where("state IN ? AND updated_at < ?", []string{string(Init), string(Accepted)}, tAccepted).
		Or("state IN ? AND updated_at < ?", []string{string(Error), string(NoAvailableWorkers)}, tReschedulable).
		Model(&Job{}).
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&jj).Error

	return
}

func (s *GormStore) Status() ([]StatusQuery, error) {
	var res []StatusQuery
	err := s.db.Debug().Raw("SELECT state, COUNT(*) as count FROM jobs GROUP BY state").Scan(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}
