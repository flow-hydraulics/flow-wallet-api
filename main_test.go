package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/eqlabs/flow-wallet-service/datastore/gorm"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/handlers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/keys/simple"
	"github.com/eqlabs/flow-wallet-service/tokens"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
)

const testDbDSN = "test.db"
const testDbType = "sqlite"

var cfg testConfig
var logger *log.Logger

type testConfig struct {
	FlowGateway string `env:"FLOW_GATEWAY,required"`
}

type TestLogger struct {
	t *testing.T
}

func (tl *TestLogger) Write(p []byte) (n int, err error) {
	tl.t.Log(fmt.Sprintf("%s", p))
	return len(p), nil
}

func TestMain(m *testing.M) {
	godotenv.Load(".env.test")

	os.Setenv("DB_DSN", testDbDSN)
	os.Setenv("DB_TYPE", testDbType)

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	logger = log.New(io.Discard, "", log.LstdFlags)

	exitcode := m.Run()

	os.Exit(exitcode)
}

func TestAccountServices(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testDbDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)

	km := simple.NewKeyManager(logger, keyStore, fc)

	wp := jobs.NewWorkerPool(logger, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	service := accounts.NewService(logger, accountStore, km, fc, wp)

	t.Run("sync create", func(t *testing.T) {
		account, err := service.CreateSync(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if err := service.ValidateAddress(account.Address); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}
	})

	t.Run("async create", func(t *testing.T) {
		job, err := service.CreateAsync()
		if err != nil {
			t.Fatal(err)
		}

		if job.Status != jobs.Accepted && job.Status != jobs.Complete {
			t.Errorf("expected job status to be %s or %s but got %s",
				jobs.Accepted, jobs.Complete, job.Status)
		}

		for job.Status == jobs.Accepted {
			time.Sleep(10 * time.Millisecond)
		}

		if job.Status != jobs.Complete {
			t.Errorf("expected job status to be %s got %s", jobs.Complete, job.Status)
		}

		account, err := service.Details(job.Result)
		if err != nil {
			t.Fatal(err)
		}

		if err := service.ValidateAddress(account.Address); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Error("Account should not expose keys")
		}
	})

	t.Run("async create thrice", func(t *testing.T) {
		_, err1 := service.CreateAsync() // Goes immediately to processing
		_, err2 := service.CreateAsync() // Queues - queue now full
		_, err3 := service.CreateAsync() // Should not fit
		if err1 != nil {
			t.Error(err1)
		}
		if err2 != nil {
			t.Error(err2)
		}
		if err3 == nil {
			t.Error("expected 503 'max capacity reached, try again later' but got no error")
		}
	})

	// Sleep a moment to allow job queue to empty
	time.Sleep(100 * time.Millisecond)

	t.Run("account can make a transaction", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		job, err := service.CreateAsync()
		if err != nil {
			t.Fatal(err)
		}

		// Wait for the job to complete
		for job.Status == jobs.Accepted {
			time.Sleep(10 * time.Millisecond)
		}

		// Fund the account from service account
		txId, err := tokens.TransferFlow(
			ctx,
			km,
			fc,
			flow.HexToAddress(job.Result),
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			"1.0",
		)
		if err != nil {
			t.Fatal(err)
		}
		_, err = flow_helpers.WaitForSeal(context.Background(), fc, txId)
		if err != nil {
			t.Fatal(err)
		}

		txId, err = tokens.TransferFlow(
			ctx,
			km,
			fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(job.Result),
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		if txId == flow.EmptyID {
			t.Fatalf("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), fc, txId)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("account can not make a transaction without funds", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		job, err := service.CreateAsync()
		if err != nil {
			t.Fatal(err)
		}

		// Wait for the job to complete
		for job.Status == jobs.Accepted {
			time.Sleep(10 * time.Millisecond)
		}

		txId, err := tokens.TransferFlow(
			ctx,
			km,
			fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(job.Result),
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		if txId == flow.EmptyID {
			t.Fatal("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), fc, txId)
		if err == nil {
			t.Fatal("Expected an error")
		}
	})
}

func TestAccountHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testDbDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	keyStore := keys.NewGormStore(db)

	km := simple.NewKeyManager(logger, keyStore, fc)

	wp := jobs.NewWorkerPool(logger, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	store := accounts.NewGormStore(db)
	service := accounts.NewService(logger, store, km, fc, wp)
	handlers := handlers.NewAccounts(logger, service)

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.List).Methods(http.MethodGet)
	router.HandleFunc("/", handlers.Create).Methods(http.MethodPost)
	router.HandleFunc("/{address}", handlers.Details).Methods(http.MethodGet)

	var tempAccAddress string

	// NOTE: The order of the test "steps" matters
	steps := []struct {
		name     string
		method   string
		url      string
		expected string
		status   int
	}{
		{
			name:     "HTTP GET accounts.List db empty",
			method:   http.MethodGet,
			url:      "/",
			expected: `\[\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP POST accounts.Create",
			method:   http.MethodPost,
			url:      "/",
			expected: `\{"jobId":".*","status":"Accepted","result":".*","createdAt":".*","updatedAt":".*"\}\n`,
			status:   http.StatusCreated,
		},
		{
			name:     "HTTP GET accounts.List db not empty",
			method:   http.MethodGet,
			url:      "/",
			expected: `\[\{"address":".*","createdAt":".*","updatedAt":".*"\}\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET accounts.Details invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address",
			expected: `not a valid address\n`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP GET accounts.Details unknown address",
			method:   http.MethodGet,
			url:      "/0f7025fa05b578e3",
			expected: `account not found\n`,
			status:   http.StatusNotFound,
		},
		{
			name:     "HTTP GET accounts.Details known address",
			method:   http.MethodGet,
			url:      "/<address>",
			expected: `\{"address":".*","createdAt":".*","updatedAt":".*"\}\n`,
			status:   http.StatusOK,
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			replacer := strings.NewReplacer(
				"<address>", tempAccAddress,
			)

			url := replacer.Replace(string(step.url))

			req, err := http.NewRequest(step.method, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Context()

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
			}

			// If this step was creating a new account
			// wait for the account to become available
			// and store the new account in "tempAcc".
			if step.status == http.StatusCreated {
				jobService := jobs.NewService(logger, jobStore)
				var rJob jobs.Job
				json.Unmarshal(rr.Body.Bytes(), &rJob)
				id := rJob.ID.String()
				job, _ := jobService.Details(id)
				for job.Status == jobs.Accepted {
					job, _ = jobService.Details(id)
				}
				tempAccAddress = job.Result
			}

			// Check the response body is what we expect.
			re := regexp.MustCompile(step.expected)
			match := re.FindString(rr.Body.String())
			if match == "" || match != rr.Body.String() {
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}
}

func TestTransactionHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	// logger = log.New(&TestLogger{t}, "", log.Lshortfile)

	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testDbDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	keyStore := keys.NewGormStore(db)

	km := simple.NewKeyManager(logger, keyStore, fc)

	wp := jobs.NewWorkerPool(logger, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	store := transactions.NewGormStore(db)
	service := transactions.NewService(logger, store, km, fc, wp)
	handlers := handlers.NewTransactions(logger, service)

	router := mux.NewRouter()
	router.HandleFunc("/{address}/transactions", handlers.List).Methods(http.MethodGet)
	router.HandleFunc("/{address}/transactions", handlers.Create).Methods(http.MethodPost)
	router.HandleFunc("/{address}/transactions/{transactionId}", handlers.Details).Methods(http.MethodGet)

	type testArg struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}

	type testTx struct {
		Code      string    `json:"code"`
		Arguments []testArg `json:"arguments"`
	}

	greeting, err := json.Marshal(&testTx{Code: `
	transaction(greeting: String) {
		execute {
			log(greeting.concat(", World!"))
		}
	}
	`,
		Arguments: []testArg{{Type: "String", Value: "Hello"}}})
	if err != nil {
		t.Fatal(err)
	}

	var tempTxId string

	// NOTE: The order of the test "steps" matters
	steps := []struct {
		name     string
		method   string
		body     io.Reader
		url      string
		expected string
		status   int
	}{
		{
			name:     "HTTP GET list db empty",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `\[\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP POST list empty body",
			method:   http.MethodPost,
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: "empty body\n",
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP POST list invalid body",
			method:   http.MethodPost,
			body:     strings.NewReader("notvalidobject"),
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: "invalid body\n",
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP POST list invalid code",
			method:   http.MethodPost,
			body:     strings.NewReader(`{"code":"test","arguments":[{"type":"string","value":"test"}]}`),
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `.*Parsing failed.*`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP POST list ok",
			method:   http.MethodPost,
			body:     strings.NewReader(string(greeting)),
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `.*"transactionId".*`,
			status:   http.StatusCreated,
		},
		{
			name:     "HTTP GET list db not empty",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `\[.*"transactionId".*\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET details invalid id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/invalid-id",
			expected: `not a valid transaction id\n`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP GET details unknown id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/unknown-id",
			expected: `transaction not found\n`,
			status:   http.StatusNotFound,
		},
		{
			name:     "HTTP GET details known id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/<id>",
			expected: `.*"transactionId".*`,
			status:   http.StatusOK,
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			replacer := strings.NewReplacer(
				"<id>", tempTxId,
			)

			url := replacer.Replace(string(step.url))

			req, err := http.NewRequest(step.method, url, step.body)
			if err != nil {
				t.Fatal(err)
			}

			req.Context()

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
			}

			// If this step was creating a new transaction store the new txId in "tempTxId".
			if step.status == http.StatusCreated {
				var transaction transactions.Transaction
				json.Unmarshal(rr.Body.Bytes(), &transaction)
				tempTxId = transaction.TransactionId
			}

			// Check the response body is what we expect.
			re := regexp.MustCompile(step.expected)
			match := re.FindString(rr.Body.String())
			if match == "" {
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}
}
