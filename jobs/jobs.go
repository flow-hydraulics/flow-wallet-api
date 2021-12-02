package jobs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Result struct {
	Result        string
	TransactionID string
}

// State is a type for Job state.
type State string

const (
	Init               State = "INIT"
	Accepted           State = "ACCEPTED"
	NoAvailableWorkers State = "NO_AVAILABLE_WORKERS"
	Error              State = "ERROR"
	Complete           State = "COMPLETE"
	Failed             State = "FAILED"
)

// Job database model
type Job struct {
	ID                     uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Type                   string         `gorm:"column:type"`
	State                  State          `gorm:"column:state;default:INIT"`
	Error                  string         `gorm:"column:error"`
	Result                 string         `gorm:"column:result"`
	TransactionID          string         `gorm:"column:transaction_id"`
	ExecCount              int            `gorm:"column:exec_count;default:0"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
	ShouldSendNotification bool           `gorm:"-"` // Whether or not to notify admin (via webhook for example)
}

type JobQueueStatus struct {
	JobsInit        int `json:"jobsInit"`
	JobsNotAccepted int `json:"jobsNotAccepted"`
	JobsAccepted    int `json:"jobsAccepted"`
	JobsErrored     int `json:"jobsErrored"`
	JobsFailed      int `json:"jobsFailed"`
	JobsCompleted   int `json:"jobsCompleted"`
}

// Job HTTP response
type JSONResponse struct {
	ID            uuid.UUID `json:"jobId"`
	Type          string    `json:"type"`
	State         State     `json:"state"`
	Error         string    `json:"error"`
	Result        string    `json:"result"`
	TransactionID string    `json:"transactionId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (j Job) ToJSONResponse() JSONResponse {
	return JSONResponse{
		ID:            j.ID,
		Type:          j.Type,
		State:         j.State,
		Error:         j.Error,
		Result:        j.Result,
		TransactionID: j.TransactionID,
		CreatedAt:     j.CreatedAt,
		UpdatedAt:     j.UpdatedAt,
	}
}

func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = uuid.New()
	return nil
}

func (j *Job) Wait(wait bool) error {
	if wait {
		// Wait for the job to have finished
		for j.State == Accepted {
			time.Sleep(10 * time.Millisecond)
		}
		if j.State == Error {
			return fmt.Errorf(j.Error)
		}
	}
	return nil
}

func (j *Job) logEntry(entry *log.Entry) *log.Entry {
	jobFields := log.Fields{
		"jobID":   j.ID,
		"jobType": j.Type,
	}

	if entry != nil {
		return entry.WithFields(jobFields)
	}

	return log.WithFields(jobFields)
}
