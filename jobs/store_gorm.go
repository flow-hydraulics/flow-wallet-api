package jobs

import (
	"fmt"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/lib"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
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

func isAcceptable(j *Job, acceptedGracePeriod time.Duration) bool {
	tAccepted := time.Now().Add(-1 * acceptedGracePeriod)
	if j.State == Accepted && j.UpdatedAt.After(tAccepted) {
		return false
	}
	if j.State == Complete || j.State == Failed {
		return false
	}
	return true
}

func (s *GormStore) AcceptJob(j *Job, acceptedGracePeriod time.Duration) error {
	if !isAcceptable(j, acceptedGracePeriod) {
		return fmt.Errorf("error job is not acceptable")
	}
	return lib.GormTransaction(s.db, func(tx *gorm.DB) error {
		var job Job
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, "id = ?", j.ID).Error
		if err != nil {
			return err
		}
		if !isAcceptable(&job, acceptedGracePeriod) {
			return fmt.Errorf("error job is not acceptable")
		}
		j.State = Accepted
		j.ExecCount = job.ExecCount + 1
		err = tx.Save(j).Error
		if err != nil {
			return err
		}
		return nil
	})
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
	err := s.db.Raw("SELECT state, COUNT(*) as count FROM jobs GROUP BY state").Scan(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}
