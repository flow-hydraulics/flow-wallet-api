package jobs

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
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
	State                  State          `gorm:"column:state;default:INIT;index:idx_jobs_state_updated_at"`
	Error                  string         `gorm:"column:error"`
	Errors                 pq.StringArray `gorm:"column:errors;type:text[]"`
	Result                 string         `gorm:"column:result"`
	TransactionID          string         `gorm:"column:transaction_id"`
	ExecCount              int            `gorm:"column:exec_count;default:0"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;index:idx_jobs_state_updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
	ShouldSendNotification bool           `gorm:"-"` // Whether or not to notify admin (via webhook for example)
	Attributes             datatypes.JSON `gorm:"attributes"`
}

func (Job) TableName() string {
	return "jobs"
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
	Errors        []string  `json:"errors"`
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
		Errors:        []string(j.Errors),
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
