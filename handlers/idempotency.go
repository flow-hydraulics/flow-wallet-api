package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Idempotency Handler middleware
// ===========================================================================

type IdempotencyStoreType int

const (
	IdempotencyStoreTypeLocal IdempotencyStoreType = iota
	IdempotencyStoreTypeShared
	IdempotencyStoreTypeRedis
)

func (ist IdempotencyStoreType) String() string {
	return [...]string{"local", "shared", "redis"}[ist]
}

type IdempotencyHandlerOptions struct {
	IgnorePaths []string
	Expiry      time.Duration
}

type IdempotencyStore interface {
	Get(key string) (bool, error)               // Get key by name, return bool "found" and possible error
	Set(key string, expiry time.Duration) error // Set key by name & expiry in seconds, return possible error
}

// Redis store for idempotency keys
type IdempotencyStoreRedis struct {
	conn   redis.Conn
	prefix string
}

func NewIdempotencyStoreRedis(c redis.Conn) *IdempotencyStoreRedis {
	return &IdempotencyStoreRedis{conn: c, prefix: "idempotencykey"}
}

func (r *IdempotencyStoreRedis) prefixedKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

func (r *IdempotencyStoreRedis) Get(key string) (bool, error) {
	keyExists, err := redis.Bool(r.conn.Do("EXISTS", r.prefixedKey(key)))
	return keyExists, err
}

func (r *IdempotencyStoreRedis) Set(key string, expiry time.Duration) error {
	res, err := r.conn.Do("PSETEX", r.prefixedKey(key), int(expiry.Milliseconds()), 1)
	if err != nil {
		return err
	}

	if res != "OK" {
		return fmt.Errorf("failed to set key: %v", res)
	}

	return nil
}

// Gorm (SQL) store for idempotency keys
type IdempotencyStoreGorm struct {
	db *gorm.DB
}

// TODO: automatically expire/prune keys w/ ExpiryDate in the past
type IdempotencyStoreGormItem struct {
	Key        string    `gorm:"column:key;primary_key"`
	ExpiryDate time.Time `gorm:"column:expiry_date"`
}

func (IdempotencyStoreGormItem) TableName() string {
	return "idempotency_keys"
}

func NewIdempotencyStoreGorm(db *gorm.DB) *IdempotencyStoreGorm {
	return &IdempotencyStoreGorm{db: db}
}

func (g *IdempotencyStoreGorm) Get(key string) (bool, error) {
	item := IdempotencyStoreGormItem{}
	err := g.db.First(&item, "key = ? and expiry_date > ?", key, time.Now()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// key doesn't exist
		return false, nil
	} else if err != nil {
		// some other error
		return false, err
	}

	// key exists
	return true, nil
}

func (g *IdempotencyStoreGorm) Set(key string, expiry time.Duration) error {
	// update expiry date if exists or create a new item
	err := g.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"expiry_date"}),
	}).Create(&IdempotencyStoreGormItem{Key: key, ExpiryDate: time.Now().Add(expiry)}).Error

	if err != nil {
		return err
	}

	return nil
}

// Prune deletes all expired IdempotencyStoreGormItems from the database
func (g *IdempotencyStoreGorm) Prune() error {
	err := g.db.Delete(IdempotencyStoreGormItem{}, "expiry_date < ?", time.Now()).Error
	return err
}

// Local / in-memory store for idempotency keys, mainly for testing purposes
type IdempotencyStoreLocal struct {
	keys map[string]time.Time // key: expiry
}

func NewIdempotencyStoreLocal() *IdempotencyStoreLocal {
	return &IdempotencyStoreLocal{make(map[string]time.Time)}
}

func (m *IdempotencyStoreLocal) Get(key string) (bool, error) {
	v, ok := m.keys[key]
	if !ok {
		return false, nil
	}

	// Still valid
	if v.After(time.Now()) {
		return true, nil
	}

	// Expired
	// NOTE: item is removed as a side effect
	if v.Before(time.Now()) {
		delete(m.keys, key)
		return false, nil
	}

	return false, nil
}

func (m *IdempotencyStoreLocal) Set(key string, expiry time.Duration) error {
	dl := time.Now().Add(expiry)
	m.keys[key] = dl

	return nil
}

// IdempotencyHandler returns a http.HandlerFunc that checks
// for request idempotency when applicable
func IdempotencyHandler(h http.Handler, opts IdempotencyHandlerOptions, store IdempotencyStore) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Check for ignored paths
		for _, path := range opts.IgnorePaths {
			if strings.HasPrefix(r.URL.Path, path) {
				h.ServeHTTP(rw, r)
				return
			}
		}

		// Only POST requests are checked
		if r.Method != http.MethodPost {
			h.ServeHTTP(rw, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if len(key) == 0 && r.Method == http.MethodPost {
			http.Error(rw, "Idempotency-Key header not found", http.StatusBadRequest)
			return
		}

		exists, err := store.Get(key)
		if err != nil {
			log.
				WithFields(log.Fields{"error": err, "key": key}).
				Warn("Error while reading idempotency key from storage")
			http.Error(rw, "Error while reading idempotency key", http.StatusInternalServerError)
			return
		}

		// XXX: same key w/ different payload should return 422,
		// but isn't necessarily required functionality => only the key is stored
		if exists {
			http.Error(rw, fmt.Sprintf("Idempotency-Key conflict, key: %s", key), http.StatusConflict)
			return
		} else {
			err := store.Set(key, opts.Expiry)
			if err != nil {
				log.
					WithFields(log.Fields{"error": err, "key": key}).
					Warn("Error while saving used idempotency key")
				http.Error(rw, "Error while saving used idempotency key", http.StatusInternalServerError)
				return
			}
		}

		h.ServeHTTP(rw, r)
	})
}
