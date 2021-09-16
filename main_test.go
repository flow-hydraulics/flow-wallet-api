package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/handlers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/gorilla/mux"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
)

const (
	testDbType            = "sqlite"
	testCadenceTxBasePath = "./cadence/transactions"
)

var testLogger *log.Logger

type TestLogger struct {
	t *testing.T
}

func (tl *TestLogger) Write(p []byte) (n int, err error) {
	tl.t.Log(string(p))
	return len(p), nil
}

type httpTestStep struct {
	name        string
	method      string
	body        io.Reader
	contentType string
	url         string
	expected    string
	status      int
	sync        bool
}

func handleStepRequest(s httpTestStep, r *mux.Router, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest(s.method, s.url, s.body)
	if err != nil {
		t.Fatal(err)
	}

	if s.contentType != "" {
		req.Header.Set("content-type", "application/json")
	}

	if s.sync {
		q := req.URL.Query()
		q.Add(handlers.SyncQueryParameter, "go-ahead")
		req.URL.RawQuery = q.Encode()
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	status := rr.Code
	// Check the status code is what we expect.
	if status != s.status {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, s.status)
	}

	// Check the response body is what we expect.
	re := regexp.MustCompile(s.expected)
	match := re.FindString(rr.Body.String())
	if match == "" {
		t.Errorf(`handler returned unexpected body: got "%s" want "%v"`, rr.Body.String(), re)
	}

	return rr
}

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var testDbDSNcounter uint = 0

func getTestDbDSN() string {
	testDbDSNcounter += 1
	return fmt.Sprintf("testDB%d.db", testDbDSNcounter)
}

func getTestConfig() *configs.Config {
	opts := &configs.Options{EnvFilePath: ".env.test"}
	var err error
	cfg, err := configs.ParseConfig(opts)
	if err != nil {
		panic(err)
	}
	cfg.DatabaseDSN = getTestDbDSN()
	cfg.DatabaseType = testDbType
	cfg.ChainID = flow.Emulator
	return cfg
}

func TestMain(m *testing.M) {
	testLogger = log.New(io.Discard, "", log.LstdFlags)

	exitcode := m.Run()

	os.Exit(exitcode)
}

func TestAccountServices(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	txStore := transactions.NewGormStore(db)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	txService := transactions.NewService(cfg, txStore, km, fc, wp)
	service := accounts.NewService(cfg, accountStore, km, fc, wp, txService)

	t.Run("admin init", func(t *testing.T) {
		ctx := context.Background()

		err = service.InitAdminAccount(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// make sure all requested proposal keys are created

		keyCount, err := km.InitAdminProposalKeys(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if keyCount != cfg.AdminProposalKeyCount {
			t.Fatal("incorrect number of admin proposal keys")
		}
	})

	t.Run("sync create", func(t *testing.T) {
		_, account, err := service.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := flow_helpers.ValidateAddress(account.Address, flow.Emulator); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}
	})

	t.Run("async create", func(t *testing.T) {
		job, _, err := service.Create(context.Background(), false)
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

		if _, err := flow_helpers.ValidateAddress(account.Address, flow.Emulator); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Error("Account should not expose keys")
		}
	})

	t.Run("async create thrice", func(t *testing.T) {
		_, _, err1 := service.Create(context.Background(), false) // Goes immediately to processing
		_, _, err2 := service.Create(context.Background(), false) // Queues - queue now full
		_, _, err3 := service.Create(context.Background(), false) // Should not fit, sometimes does
		if err1 != nil {
			t.Error(err1)
		}
		if err2 != nil {
			t.Error(err2)
		}
		if err3 == nil {
			t.Log("expected 503 'max capacity reached, try again later' but got no error")
		}
	})
}

func TestAccountHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	txStore := transactions.NewGormStore(db)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	store := accounts.NewGormStore(db)
	txService := transactions.NewService(cfg, txStore, km, fc, wp)
	service := accounts.NewService(cfg, store, km, fc, wp, txService)

	h := handlers.NewAccounts(testLogger, service)

	t.Run("admin init", func(t *testing.T) {
		err = service.InitAdminAccount(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	router := mux.NewRouter()
	router.Handle("/", h.List()).Methods(http.MethodGet)
	router.Handle("/", h.Create()).Methods(http.MethodPost)
	router.Handle("/{address}", h.Details()).Methods(http.MethodGet)

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
			name:     "list db empty",
			method:   http.MethodGet,
			url:      "/",
			expected: fmt.Sprintf(`(?m)^\[{"address":"%s".*}\]$`, cfg.AdminAddress),
			status:   http.StatusOK,
		},
		{
			name:     "create",
			method:   http.MethodPost,
			url:      "/",
			expected: `(?m)^{"jobId":".+"}$`,
			status:   http.StatusCreated,
		},
		{
			name:     "list db not empty",
			method:   http.MethodGet,
			url:      "/",
			expected: `(?m)^\[{"address":".+".*}\]$`,
			status:   http.StatusOK,
		},
		{
			name:     "details invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address",
			expected: "not a valid address",
			status:   http.StatusBadRequest,
		},
		{
			name:     "details unknown address",
			method:   http.MethodGet,
			url:      "/0f7025fa05b578e3",
			expected: "record not found",
			status:   http.StatusNotFound,
		},
		{
			name:     "details known address",
			method:   http.MethodGet,
			url:      "/<address>",
			expected: `(?m)^{"address":".+"}$`,
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
				jobService := jobs.NewService(jobStore)
				var rJob jobs.JSONResponse
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
			if match == "" {
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}
}

func TestAccountTransactionHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	// logger = log.New(&TestLogger{t}, "", log.Lshortfile)

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	keyStore := keys.NewGormStore(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	store := transactions.NewGormStore(db)
	service := transactions.NewService(cfg, store, km, fc, wp)
	h := handlers.NewTransactions(testLogger, service)

	router := mux.NewRouter()
	router.Handle("/{address}/transactions", h.List()).Methods(http.MethodGet)
	router.Handle("/{address}/transactions", h.Create()).Methods(http.MethodPost)
	router.Handle("/{address}/transactions/{transactionId}", h.Details()).Methods(http.MethodGet)

	token, err := templateService.GetTokenByName("FlowToken")
	if err != nil {
		t.Fatal(err)
	}

	tFlow := templates.FungibleTransferCode(cfg.ChainID, token)
	tFlowBytes, err := json.Marshal(tFlow)
	if err != nil {
		t.Fatal(err)
	}

	validTransferFlow := fmt.Sprintf(`{
		"code":%s,
		"arguments":[{"type":"UFix64","value":"1.0"},{"type":"Address","value":"%s"}]
	}`, tFlowBytes, cfg.AdminAddress)

	validHello := `{
		"code":"transaction(greeting: String) { prepare(signer: AuthAccount){} execute { log(greeting.concat(\", World!\")) }}",
		"arguments":[{"type":"String","value":"Hello"}]
	}`

	invalidHello := `{
		"code":"this is not valid cadence code",
		"arguments":[{"type":"String","value":"Hello"}]
	}`

	var tempTxId string

	// NOTE: The order of the test "steps" matters
	steps := []struct {
		name        string
		method      string
		body        io.Reader
		contentType string
		url         string
		expected    string
		status      int
		sync        bool
	}{
		{
			name:     "list db empty",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected: `(?m)^\[\]$`,
			status:   http.StatusOK,
		},
		{
			name:     "list db empty invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address/transactions",
			expected: "not a valid address",
			status:   http.StatusBadRequest,
		},
		{
			name:        "create ok async",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    `(?m)^{"jobId":".+"}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "create ok sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    `(?m)^{"transactionId":".+"}$`,
			status:      http.StatusCreated,
			sync:        true,
		},
		{
			name:        "create invalid content-type",
			method:      http.MethodPost,
			contentType: "",
			body:        strings.NewReader(validHello),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    "Unsupported content type",
			status:      http.StatusUnsupportedMediaType,
			sync:        true,
		},
		{
			name:        "create ok sync requires authorizer",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validTransferFlow),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    `(?m)^{"transactionId":".+"}$`,
			status:      http.StatusCreated,
			sync:        true,
		},
		{
			name:        "create nil body sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        nil,
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    "empty body",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "create empty body sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(""),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    "empty body",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "create invalid body sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader("notvalidobject"),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    "invalid body",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "create invalid code sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(invalidHello),
			url:         fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected:    "Parsing failed",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "create invalid address sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         "/invalid-address/transactions",
			expected:    "not a valid address",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:     "list db not empty",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s/transactions", cfg.AdminAddress),
			expected: `(?m)^\[{"transactionId":".+".*}\]$`,
			status:   http.StatusOK,
		},
		{
			name:     "details invalid id",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s/transactions/invalid-id", cfg.AdminAddress),
			expected: "not a valid transaction id",
			status:   http.StatusBadRequest,
		},
		{
			name:     "details unknown id",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s/transactions/0e4f500d6965c7fc0ff1239525e09eb9dd27c00a511976e353d9f6a44ca22921", cfg.AdminAddress),
			expected: "transaction not found",
			status:   http.StatusNotFound,
		},
		{
			name:     "details known id",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s/transactions/<id>", cfg.AdminAddress),
			expected: `(?m)^{"transactionId":".+"}$`,
			status:   http.StatusOK,
		},
		{
			name:     "details invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address/transactions/invalid-id",
			expected: "not a valid address",
			status:   http.StatusBadRequest,
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

			if step.contentType != "" {
				req.Header.Set("content-type", "application/json")
			}

			if step.sync {
				q := req.URL.Query()
				q.Add(handlers.SyncQueryParameter, "go-ahead")
				req.URL.RawQuery = q.Encode()
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			status := rr.Code
			// Check the status code is what we expect.
			if status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
			}

			// If this step was creating a new transaction store the new txId in "tempTxId".
			if step.sync && status == http.StatusCreated {
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

func TestTransactionHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	keyStore := keys.NewGormStore(db)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	transactionStore := transactions.NewGormStore(db)
	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)

	accountStore := accounts.NewGormStore(db)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService)

	h := handlers.NewTransactions(testLogger, transactionService)

	router := mux.NewRouter()
	router.Handle("/transactions", h.List()).Methods(http.MethodGet)
	router.Handle("/transactions/{transactionId}", h.Details()).Methods(http.MethodGet)

	err = accountService.InitAdminAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	token, err := templateService.GetTokenByName("FlowToken")
	if err != nil {
		t.Fatal(err)
	}

	transferFlow := templates.FungibleTransferCode(cfg.ChainID, token)

	_, transaction1, err := transactionService.Create(
		context.Background(),
		true,
		cfg.AdminAddress,
		templates.Raw{
			Code:      "transaction() { prepare(signer: AuthAccount){} execute { log(\"Hello World!\") }}",
			Arguments: []templates.Argument{},
		},
		transactions.General,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, transaction2, err := transactionService.Create(
		context.Background(),
		true,
		cfg.AdminAddress,
		templates.Raw{
			Code: transferFlow,
			Arguments: []templates.Argument{
				cadence.UFix64(1.0),
				cadence.NewAddress(flow.HexToAddress(cfg.AdminAddress)),
			},
		},
		transactions.General,
	)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: The order of the test "steps" matters
	steps := []struct {
		name        string
		method      string
		body        io.Reader
		contentType string
		url         string
		expected    string
		status      int
		sync        bool
	}{
		{
			name:     "list db not empty",
			method:   http.MethodGet,
			url:      "/transactions",
			expected: `(?m)^\[{\"transactionId\":.*\]$`,
			status:   http.StatusOK,
		},
		{
			name:     "details invalid id",
			method:   http.MethodGet,
			url:      "/transactions/invalid-id",
			expected: "not a valid transaction id",
			status:   http.StatusBadRequest,
		},
		{
			name:     "details unknown id",
			method:   http.MethodGet,
			url:      "/transactions/0e4f500d6965c7fc0ff1239525e09eb9dd27c00a511976e353d9f6a44ca22921",
			expected: "transaction not found",
			status:   http.StatusNotFound,
		},
		{
			name:     "details",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/transactions/%s", transaction1.TransactionId),
			expected: `(?m)^{"transactionId":"\w+".*}$`,
			status:   http.StatusOK,
		},
		{
			name:     "details with events",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/transactions/%s", transaction2.TransactionId),
			expected: `(?m)^{"transactionId":"\w+".*"events":.*}$`,
			status:   http.StatusOK,
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			req, err := http.NewRequest(step.method, step.url, step.body)
			if err != nil {
				t.Fatal(err)
			}

			if step.contentType != "" {
				req.Header.Set("content-type", "application/json")
			}

			if step.sync {
				q := req.URL.Query()
				q.Add(handlers.SyncQueryParameter, "go-ahead")
				req.URL.RawQuery = q.Encode()
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			status := rr.Code
			// Check the status code is what we expect.
			if status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
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

func TestScriptsHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	service := transactions.NewService(cfg, nil, nil, fc, nil)
	h := handlers.NewTransactions(testLogger, service)

	router := mux.NewRouter()
	router.Handle("/", h.ExecuteScript()).Methods(http.MethodPost)

	steps := []struct {
		name        string
		method      string
		body        io.Reader
		contentType string
		expected    string
		status      int
	}{
		{
			name:   "int 1",
			method: http.MethodPost,
			body: strings.NewReader(`{
				"code":"pub fun main(): Int { return 1 }",
				"arguments":[]
			}`),
			contentType: "application/json",
			expected:    `(?m)^{"Value":1}$`,
			status:      http.StatusOK,
		},
		{
			name:   "get supply",
			method: http.MethodPost,
			body: strings.NewReader(`{
				"code":"import FlowToken from 0x0ae53cb6e3f42a79\npub fun main(): UFix64 {\nlet supply = FlowToken.totalSupply\nreturn supply\n}",
				"arguments":[]
			}`),
			contentType: "application/json",
			expected:    "100000000000000000",
			status:      http.StatusOK,
		},
		{
			name:   "get balance",
			method: http.MethodPost,
			body: strings.NewReader(fmt.Sprintf(`{
				"code":"import FungibleToken from 0xee82856bf20e2aa6\nimport FlowToken from 0x0ae53cb6e3f42a79\npub fun main(account: Address): UFix64 {\nlet vaultRef = getAccount(account)\n.getCapability(/public/flowTokenBalance)\n.borrow<&FlowToken.Vault{FungibleToken.Balance}>()\n?? panic(\"Could not borrow Balance reference to the Vault\")\nreturn vaultRef.balance\n}",
				"arguments":[{"type":"Address","value":"%s"}]
			}`, cfg.AdminAddress)),
			contentType: "application/json",
			expected:    "\\d+",
			status:      http.StatusOK,
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			req, err := http.NewRequest(step.method, "/", step.body)
			if err != nil {
				t.Fatal(err)
			}

			if step.contentType != "" {
				req.Header.Set("content-type", "application/json")
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			status := rr.Code
			// Check the status code is what we expect.
			if status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
			}

			// Check the response body is what we expect.
			re := regexp.MustCompile(step.expected)
			match := re.FindString(rr.Body.String())
			if match == "" {
				t.Errorf("handler returned unexpected body: got %s want %v", rr.Body.String(), re)
			}
		})
	}
}

func TestTokenServices(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService)
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, transactionService, templateService, accountService)

	t.Run("admin init", func(t *testing.T) {
		err = accountService.InitAdminAccount(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("account can make a transaction", func(t *testing.T) {
		// Create an account
		_, account, err := accountService.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		// Fund the account from service account
		_, _, err = tokenService.CreateWithdrawal(
			context.Background(),
			true,
			cfg.AdminAddress,
			tokens.WithdrawalRequest{
				TokenName: "FlowToken",
				Recipient: account.Address,
				FtAmount:  "1.0",
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		_, transfer, err := tokenService.CreateWithdrawal(
			context.Background(),
			true,
			account.Address,
			tokens.WithdrawalRequest{
				TokenName: "FlowToken",
				Recipient: cfg.AdminAddress,
				FtAmount:  "1.0",
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if flow.HexToID(transfer.TransactionId) == flow.EmptyID {
			t.Fatalf("Expected TransactionId not to be empty")
		}
	})

	t.Run("account can not make a transaction without funds", func(t *testing.T) {
		// Create an account
		_, account, err := accountService.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		_, tx, err := tokenService.CreateWithdrawal(
			context.Background(),
			true,
			account.Address,
			tokens.WithdrawalRequest{
				TokenName: "FlowToken",
				Recipient: cfg.AdminAddress,
				FtAmount:  "1.0",
			},
		)

		if tx == nil {
			t.Fatal("Expected transaction not to be nil")
		}

		if flow.HexToID(tx.TransactionId) == flow.EmptyID {
			t.Fatal("Expected TransactionId not to be empty")
		}

		if err == nil {
			t.Fatal("Expected an error")
		}
	})

	t.Run("manage fusd for an account", func(t *testing.T) {
		tokenName := "FUSD"

		ctx := context.Background()

		// Make sure FUSD is deployed
		_, _, err := tokenService.DeployTokenContractForAccount(ctx, true, tokenName, cfg.AdminAddress)
		if err != nil {
			if !strings.Contains(err.Error(), "cannot overwrite existing contract") {
				t.Fatal(err)
			}
		}

		// Setup the admin account to be able to handle FUSD
		_, _, err = tokenService.Setup(ctx, true, tokenName, cfg.AdminAddress)
		if err != nil {
			if !strings.Contains(err.Error(), "vault exists") {
				t.Fatal(err)
			}
		}

		// Create an account
		_, account, err := accountService.Create(ctx, true)
		if err != nil {
			t.Fatal(err)
		}

		// Setup the new account to be able to handle FUSD
		_, setupTx, err := tokenService.Setup(ctx, true, tokenName, account.Address)
		if err != nil {
			t.Fatal(err)
		}

		if setupTx.TransactionType != transactions.FtSetup {
			t.Errorf("expected %s but got %s", transactions.FtSetup, setupTx.TransactionType)
		}

		// Create a withdrawal, should error as we can not mint FUSD right now
		_, _, err = tokenService.CreateWithdrawal(
			ctx,
			true,
			cfg.AdminAddress,
			tokens.WithdrawalRequest{
				TokenName: tokenName,
				Recipient: account.Address,
				FtAmount:  "1.0",
			},
		)
		if err != nil {
			if !strings.Contains(err.Error(), "Amount withdrawn must be less than or equal than the balance of the Vault") {
				t.Fatal(err)
			}
		}

	})

	t.Run(("try to setup a non-existent token"), func(t *testing.T) {
		tokenName := "some-non-existent-token"

		ctx := context.Background()

		// Create an account
		_, account, err := accountService.Create(ctx, true)
		if err != nil {
			t.Fatal(err)
		}

		// Setup the new account to be able to handle the non-existent token
		_, _, err = tokenService.Setup(ctx, true, tokenName, account.Address)
		if err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestTokenHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService)
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, transactionService, templateService, accountService)

	t.Run("admin init", func(t *testing.T) {
		err = accountService.InitAdminAccount(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	tokenHandler := handlers.NewTokens(testLogger, tokenService)

	router := mux.NewRouter()
	router.Handle("/{address}/fungible-tokens", tokenHandler.AccountTokens(templates.FT)).Methods(http.MethodGet)
	router.Handle("/{address}/fungible-tokens/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
	router.Handle("/{address}/fungible-tokens/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
	router.Handle("/{address}/fungible-tokens/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
	router.Handle("/{address}/fungible-tokens/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
	router.Handle("/{address}/fungible-tokens/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
	router.Handle("/{address}/fungible-tokens/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
	router.Handle("/{address}/fungible-tokens/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)

	router.Handle("/{address}/non-fungible-tokens", tokenHandler.AccountTokens(templates.NFT)).Methods(http.MethodGet)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
	router.Handle("/{address}/non-fungible-tokens/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)

	// Setup

	// FlowToken
	flowToken, err := templateService.GetTokenByName("FlowToken")
	if err != nil {
		t.Fatal(err)
	}

	// FUSD
	fusd, err := templateService.GetTokenByName("FUSD")
	if err != nil {
		t.Fatal(err)
	}

	// Make sure FUSD is deployed
	_, _, err = tokenService.DeployTokenContractForAccount(context.Background(), true, fusd.Name, fusd.Address)
	if err != nil {
		if !strings.Contains(err.Error(), "cannot overwrite existing contract") {
			t.Fatal(err)
		}
	}

	// ExampleNFT

	setupBytes, err := ioutil.ReadFile(filepath.Join(testCadenceTxBasePath, "setup_exampleNFT.cdc"))
	fatal(t, err)

	transferBytes, err := ioutil.ReadFile(filepath.Join(testCadenceTxBasePath, "transfer_exampleNFT.cdc"))
	fatal(t, err)

	balanceBytes, err := ioutil.ReadFile(filepath.Join(testCadenceTxBasePath, "balance_exampleNFT.cdc"))
	fatal(t, err)

	mintBytes, err := ioutil.ReadFile(filepath.Join(testCadenceTxBasePath, "mint_exampleNFT.cdc"))
	fatal(t, err)

	exampleNft := templates.Token{
		Name:     "ExampleNFT",
		Address:  cfg.AdminAddress,
		Type:     templates.NFT,
		Setup:    string(setupBytes),
		Transfer: string(transferBytes),
		Balance:  string(balanceBytes),
	}

	// Make sure ExampleNFT is enabled
	if err := templateService.AddToken(&exampleNft); err != nil {
		t.Fatal(err)
	}

	// Make sure ExampleNFT is deployed
	_, _, err = tokenService.DeployTokenContractForAccount(context.Background(), true, exampleNft.Name, exampleNft.Address)
	if err != nil {
		if !strings.Contains(err.Error(), "cannot overwrite existing contract") {
			t.Fatal(err)
		}
	}

	// Create a few accounts
	testAccounts := make([]*accounts.Account, 2)
	for i := 0; i < 2; i++ {
		_, a, err := accountService.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}
		testAccounts[i] = a
	}

	_, testAccount, err := accountService.Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	_, testTransferFT, err := tokenService.CreateWithdrawal(
		context.Background(),
		true,
		cfg.AdminAddress,
		tokens.WithdrawalRequest{
			TokenName: flowToken.Name,
			Recipient: testAccount.Address,
			FtAmount:  "1.0",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Setup tokens
	setupTokenSteps := []httpTestStep{
		{
			name:        "Setup FlowToken async",
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[0].Address, flowToken.Name),
			expected:    `(?m)^{"jobId":".+".*}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "Setup FlowToken sync",
			sync:        true,
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[1].Address, flowToken.Name),
			expected:    `vault exists`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Setup FUSD valid async",
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[0].Address, fusd.Name),
			expected:    `(?m)^{"jobId":".+".*}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "Setup FUSD valid sync",
			sync:        true,
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[1].Address, fusd.Name),
			expected:    `(?m)^{"transactionId":".+".*}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "Setup ExampleNFT valid async",
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s", testAccounts[0].Address, exampleNft.Name),
			expected:    `(?m)^{"jobId":".+".*}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "Setup ExampleNFT valid sync",
			sync:        true,
			method:      http.MethodPost,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s", testAccounts[1].Address, exampleNft.Name),
			expected:    `(?m)^{"transactionId":".+".*}$`,
			status:      http.StatusCreated,
		},
	}

	for _, s := range setupTokenSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
			time.Sleep(100 * time.Millisecond)
		})
	}

	// Mint ExampleNFTs for account 0
	mintCode := templates.TokenCode(cfg.ChainID, &exampleNft, string(mintBytes))
	for i := 0; i < 3; i++ {
		_, _, err := transactionService.Create(context.Background(), true, cfg.AdminAddress, templates.Raw{
			Code: mintCode,
			Arguments: []templates.Argument{
				cadence.NewAddress(flow.HexToAddress(testAccounts[0].Address)),
			},
		}, transactions.General)
		fatal(t, err)
	}

	aa0NftDetails, err := tokenService.Details(context.Background(), exampleNft.Name, testAccounts[0].Address)
	fatal(t, err)

	nftIDs := aa0NftDetails.Balance.CadenceValue.(cadence.Array).Values

	_, testTransferNFT, err := tokenService.CreateWithdrawal(
		context.Background(),
		true,
		testAccounts[0].Address,
		tokens.WithdrawalRequest{
			TokenName: exampleNft.Name,
			Recipient: testAccounts[1].Address,
			NftID:     reflect.ValueOf(nftIDs[0].ToGoValue()).Uint(),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Token details
	detailtsSteps := []httpTestStep{
		{
			name:        "FlowToken details",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[1].Address, flowToken.Name),
			expected:    `(?m)^{"name":"FlowToken","balance":"\d*\.?\d*"}$`,
			status:      http.StatusOK,
		},
		{
			name:        "FUSD details",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s", testAccounts[1].Address, fusd.Name),
			expected:    `(?m)^{"name":"FUSD\","balance":"\d*\.?\d*"}$`,
			status:      http.StatusOK,
		},
		{
			name:        "ExampleNFT details",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s", testAccounts[1].Address, exampleNft.Name),
			expected:    `(?m)^{"name":"ExampleNFT\","balance":\[\d+\]}$`,
			status:      http.StatusOK,
		},
		{
			name:        "ExampleNFT details",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s", testAccounts[0].Address, exampleNft.Name),
			expected:    `(?m)^{"name":"ExampleNFT\","balance":\[(\d+,)+\d+\]}$`,
			status:      http.StatusOK,
		},
	}

	// Token list
	listSteps := []httpTestStep{
		{
			name:        "list account fungible tokens",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens", testAccounts[1].Address),
			expected:    `(?m)^\[{"name":"FUSD".*"name":"FlowToken".*}\]$`,
			status:      http.StatusOK,
		},
	}

	// Create withdrawals
	createWithdrawalSteps := []httpTestStep{
		{
			name:        "create withdrawal valid async",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":"1.0"}`, testAccount.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", cfg.AdminAddress, flowToken.Name),
			expected:    `(?m)^{"jobId":".+"}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "create withdrawal valid sync",
			sync:        true,
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":"1.0"}`, testAccount.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", cfg.AdminAddress, flowToken.Name),
			expected:    `(?m)^{"transactionId":".+"}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "create withdrawal invalid recipient",
			method:      http.MethodPost,
			body:        strings.NewReader(`{"recipient":"","amount":"1.0"}`),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", cfg.AdminAddress, flowToken.Name),
			expected:    "not a valid address",
			status:      http.StatusBadRequest,
		},
		{
			name:        "create withdrawal invalid amount",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":""}`, testAccount.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", cfg.AdminAddress, flowToken.Name),
			expected:    "missing decimal point",
			status:      http.StatusBadRequest,
		},
		{
			name:        "create ExampleNFT withdrawal valid sync",
			sync:        true,
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","nftId":%d}`, cfg.AdminAddress, nftIDs[1].ToGoValue())),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s/withdrawals", testAccounts[0].Address, exampleNft.Name),
			expected:    `(?m)^{"transactionId":".+"}$`,
			status:      http.StatusCreated,
		},
		{
			name:        "create ExampleNFT withdrawal with missing NFT",
			sync:        true,
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","nftId":%d}`, cfg.AdminAddress, nftIDs[1].ToGoValue())),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s/withdrawals", testAccounts[1].Address, exampleNft.Name),
			expected:    `missing NFT`,
			status:      http.StatusBadRequest,
		},
	}

	// List withdrawals
	listWithdrawalSteps := []httpTestStep{
		{
			name:        "list fungible token withdrawals valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", cfg.AdminAddress, flowToken.Name),
			expected:    `(?m)^\[{"transactionId":".+".*}\]$`,
			status:      http.StatusOK,
		},
		{
			name:        "list non-fungible token withdrawals valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s/withdrawals", testAccounts[0].Address, exampleNft.Name),
			expected:    `(?m)^\[{"transactionId":".+".*}\]$`,
			status:      http.StatusOK,
		},
		{
			name:        "list withdrawals empty",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals", testAccount.Address, flowToken.Name),
			expected:    `(?m)^\[\]$`,
			status:      http.StatusOK,
		},
		{
			name:        "get withdrawal details valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals/%s", cfg.AdminAddress, flowToken.Name, testTransferFT.TransactionId),
			expected:    fmt.Sprintf(`(?m)^{"transactionId":"%s".*}$`, testTransferFT.TransactionId),
			status:      http.StatusOK,
		},
	}

	// Get withdrawals
	getWithdrawalSteps := []httpTestStep{
		{
			name:        "get fungible token withdrawal details valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals/%s", cfg.AdminAddress, flowToken.Name, testTransferFT.TransactionId),
			expected:    fmt.Sprintf(`(?m)^{"transactionId":"%s".*}$`, testTransferFT.TransactionId),
			status:      http.StatusOK,
		},
		{
			name:        "get non-fungible token withdrawal details valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/non-fungible-tokens/%s/withdrawals/%s", testAccounts[0].Address, exampleNft.Name, testTransferNFT.TransactionId),
			expected:    fmt.Sprintf(`(?m)^{"transactionId":"%s".*}$`, testTransferNFT.TransactionId),
			status:      http.StatusOK,
		},
		{
			name:        "get withdrawal details invalid transaction id",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals/invalid-transaction-id", cfg.AdminAddress, flowToken.Name),
			expected:    "not a valid transaction id",
			status:      http.StatusBadRequest,
		},
		{
			name:        "get withdrawal details not found",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/withdrawals/%s", testAccount.Address, flowToken.Name, testTransferFT.TransactionId),
			expected:    "record not found",
			status:      http.StatusNotFound,
		},
	}

	// List deposits
	listDepositSteps := []httpTestStep{
		{
			name:        "list deposits valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits", testAccount.Address, flowToken.Name),
			expected:    `(?m)^\[{"transactionId":".+".*}\]$`,
			status:      http.StatusOK,
		},
		{
			name:        "list deposits invalid token name",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits", testAccount.Address, "some-invalid-token-name"),
			expected:    `record not found`,
			status:      http.StatusNotFound,
		},
		{
			name:        "list deposits invalid address",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits", "0x1", flowToken.Name),
			expected:    `not a valid address`,
			status:      http.StatusBadRequest,
		},
	}

	// Get deposits
	getDepositSteps := []httpTestStep{
		{
			name:        "get deposit details valid",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits/%s", testAccount.Address, flowToken.Name, testTransferFT.TransactionId),
			expected:    `(?m)^{"transactionId":".+".*}$`,
			status:      http.StatusOK,
		},
		{
			name:        "get deposit details invalid token name",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits/%s", testAccount.Address, "some-invalid-token-name", testTransferFT.TransactionId),
			expected:    `record not found`,
			status:      http.StatusNotFound,
		},
		{
			name:        "get deposit details invalid address",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits/%s", "0x1", flowToken.Name, testTransferFT.TransactionId),
			expected:    `not a valid address`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "get deposit details invalid transactionId",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits/%s", testAccount.Address, flowToken.Name, "0"),
			expected:    `not a valid transaction id`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "get deposit details 404",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/%s/deposits/%s", testAccounts[0].Address, flowToken.Name, testTransferFT.TransactionId),
			expected:    `record not found`,
			status:      http.StatusNotFound,
		},
	}

	for _, s := range detailtsSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range listSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range createWithdrawalSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range listWithdrawalSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range getWithdrawalSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range listDepositSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range getDepositSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}
}

func TestNFTDeployment(t *testing.T) {
	t.Skip("currently not supported")

	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	defer wp.Stop()

	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService)
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, transactionService, templateService, accountService)

	_, _, err = tokenService.DeployTokenContractForAccount(context.Background(), true, "ExampleNFT", cfg.AdminAddress)
	if err != nil {
		if !strings.Contains(err.Error(), "cannot overwrite existing contract") {
			t.Fatal(err)
		}
	}
}

func TestTemplateHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)
	templateHandler := handlers.NewTemplates(testLogger, templateService)

	router := mux.NewRouter()
	router.Handle("/tokens", templateHandler.AddToken()).Methods(http.MethodPost)
	router.Handle("/tokens", templateHandler.ListTokens(templates.NotSpecified)).Methods(http.MethodGet)
	router.Handle("/tokens/{id_or_name}", templateHandler.GetToken()).Methods(http.MethodGet)
	router.Handle("/tokens/{id}", templateHandler.RemoveToken()).Methods(http.MethodDelete)

	addStepts := []httpTestStep{
		{
			name:        "Add with empty body",
			method:      http.MethodPost,
			contentType: "application/json",
			url:         "/tokens",
			expected:    `empty body`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Add with invalid body",
			method:      http.MethodPost,
			body:        strings.NewReader(`invalid`),
			contentType: "application/json",
			url:         "/tokens",
			expected:    `invalid body`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Add with invalid name",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"","address":"%s"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    `not a valid name: ""`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Add with invalid address",
			method:      http.MethodPost,
			body:        strings.NewReader(`{"name":"TestToken","address":"0x1"}`),
			contentType: "application/json",
			url:         "/tokens",
			expected:    `not a valid address: "0x1"`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Add with valid body",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"TestToken","address":"%s"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    fmt.Sprintf(`{"id":\d+,"name":"TestToken","address":"%s","type":"NotSpecified"}`, cfg.AdminAddress),
			status:      http.StatusCreated,
		},
		{
			name:        "Add duplicate",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"TestToken","address":"%s"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    `UNIQUE constraint failed: tokens.name`,
			status:      http.StatusBadRequest,
		},
		{
			name:        "List not empty",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         "/tokens",
			expected:    fmt.Sprintf(`\[{"id":\d+,"name":"TestToken","address":"%s","type":"NotSpecified"}.*\]`, cfg.AdminAddress),
			status:      http.StatusOK,
		},
	}

	getSteps := []httpTestStep{
		{
			name:        "Get invalid id",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         "/tokens/invalid-id",
			expected:    `record not found`,
			status:      http.StatusNotFound,
		},
		{
			name:        "Get not found",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         "/tokens/100",
			expected:    `record not found`,
			status:      http.StatusNotFound,
		},
		{
			name:        "Get valid id",
			method:      http.MethodGet,
			contentType: "application/json",
			url:         "/tokens/1",
			expected:    `{"id":1,.*"type":"NotSpecified"}`,
			status:      http.StatusOK,
		},
	}

	removeSteps := []httpTestStep{
		{
			name:        "Remove invalid id",
			method:      http.MethodDelete,
			contentType: "application/json",
			url:         "/tokens/invalid-id",
			expected:    `parsing "invalid-id": invalid syntax`,
			status:      http.StatusBadRequest,
		},
		{
			// Gorm won't return an error if deleting a non-existent entry
			name:        "Remove not found",
			method:      http.MethodDelete,
			contentType: "application/json",
			url:         "/tokens/100",
			expected:    `100`,
			status:      http.StatusOK,
		},
		{
			name:        "Remove valid id",
			method:      http.MethodDelete,
			contentType: "application/json",
			url:         "/tokens/1",
			expected:    "1",
			status:      http.StatusOK,
		},
	}

	typeSteps := []httpTestStep{
		{
			name:        "Add invalid type",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"TestToken2","address":"%s","type":"not-a-valid-type"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    fmt.Sprintf(`{"id":\d+,"name":"TestToken2","address":"%s","type":"NotSpecified"}`, cfg.AdminAddress),
			status:      http.StatusCreated,
		},
		{
			name:        "Add FT valid",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"TestToken3","address":"%s","type":"FT"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    fmt.Sprintf(`{"id":\d+,"name":"TestToken3","address":"%s","type":"FT"}`, cfg.AdminAddress),
			status:      http.StatusCreated,
		},
		{
			name:        "Add NFT valid",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"name":"TestToken4","address":"%s","type":"NFT"}`, cfg.AdminAddress)),
			contentType: "application/json",
			url:         "/tokens",
			expected:    fmt.Sprintf(`{"id":\d+,"name":"TestToken4","address":"%s","type":"NFT"}`, cfg.AdminAddress),
			status:      http.StatusCreated,
		},
	}

	for _, s := range addStepts {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range getSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range removeSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}

	for _, s := range typeSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}
}

func TestTemplateService(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	cfg := getTestConfig()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfg.DatabaseDSN)
	defer gorm.Close(db)

	store := templates.NewGormStore(db)
	service := templates.NewService(cfg, store)

	// Add a token for testing
	token := templates.Token{Name: "RandomTokenName", Address: cfg.AdminAddress}
	err = service.AddToken(&token)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get token by name", func(t *testing.T) {
		t1, err := service.GetTokenByName("RandomTokenName")
		if err != nil {
			t.Fatal(err)
		}

		t2, err := service.GetTokenByName("randomtokenname")
		if err != nil {
			t.Fatal(err)
		}

		t3, err := service.GetTokenByName("randomTokenName")
		if err != nil {
			t.Fatal(err)
		}

		_, err = service.GetTokenByName("othername")
		if err == nil {
			t.Error("expected an error")
		}

		if t1.Address != token.Address || t2.Address != token.Address || t3.Address != token.Address {
			t.Error("expected tokens to be equal")
		}
	})

}
