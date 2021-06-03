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
	"github.com/eqlabs/flow-wallet-service/keys/basic"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
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
	AccessApiHost string       `env:"ACCESS_API_HOST,required"`
	ChainId       flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
}

type TestLogger struct {
	t *testing.T
}

func (tl *TestLogger) Write(p []byte) (n int, err error) {
	tl.t.Log(fmt.Sprintf("%s", p))
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
		req.Header.Set(handlers.SYNC_HEADER, "go ahead")
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
		t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
	}

	return rr
}

func TestMain(m *testing.M) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Println("WARNING: Could not load environment variables from file; ", err)
	}

	if err = os.Setenv("DATABASE_DSN", testDbDSN); err != nil {
		panic(err)
	}
	if err = os.Setenv("DATABASE_TYPE", testDbType); err != nil {
		panic(err)
	}

	if err = env.Parse(&cfg); err != nil {
		panic(err)
	}

	logger = log.New(io.Discard, "", log.LstdFlags)

	exitcode := m.Run()

	os.Exit(exitcode)
}

func TestAccountServices(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
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

	km := basic.NewKeyManager(keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	service := accounts.NewService(accountStore, km, fc, wp)

	t.Run("sync create", func(t *testing.T) {
		_, account, err := service.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		if err := flow_helpers.ValidateAddress(account.Address, flow.Emulator); err != nil {
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

		if err := flow_helpers.ValidateAddress(account.Address, flow.Emulator); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Error("Account should not expose keys")
		}
	})

	t.Run("async create thrice", func(t *testing.T) {
		_, _, err1 := service.Create(context.Background(), false) // Goes immediately to processing
		_, _, err2 := service.Create(context.Background(), false) // Queues - queue now full
		_, _, err3 := service.Create(context.Background(), false) // Should not fit
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
}

func TestAccountHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
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

	km := basic.NewKeyManager(keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	store := accounts.NewGormStore(db)
	service := accounts.NewService(store, km, fc, wp)
	h := handlers.NewAccounts(logger, service)

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
			name:     "HTTP GET accounts.List db empty",
			method:   http.MethodGet,
			url:      "/",
			expected: `\[\]`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP POST accounts.Create",
			method:   http.MethodPost,
			url:      "/",
			expected: `"jobId":".+"`,
			status:   http.StatusCreated,
		},
		{
			name:     "HTTP GET accounts.List db not empty",
			method:   http.MethodGet,
			url:      "/",
			expected: `\[.*"address":".+".*\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET accounts.Details invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address",
			expected: "not a valid address",
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP GET accounts.Details unknown address",
			method:   http.MethodGet,
			url:      "/0f7025fa05b578e3",
			expected: "account not found",
			status:   http.StatusNotFound,
		},
		{
			name:     "HTTP GET accounts.Details known address",
			method:   http.MethodGet,
			url:      "/<address>",
			expected: `"address":".+"`,
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
			if match == "" {
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}
}

func TestTransactionHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	// logger = log.New(&TestLogger{t}, "", log.Lshortfile)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
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

	km := basic.NewKeyManager(keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	store := transactions.NewGormStore(db)
	service := transactions.NewService(store, km, fc, wp)
	h := handlers.NewTransactions(logger, service)

	router := mux.NewRouter()
	router.Handle("/{address}/transactions", h.List()).Methods(http.MethodGet)
	router.Handle("/{address}/transactions", h.Create()).Methods(http.MethodPost)
	router.Handle("/{address}/transactions/{transactionId}", h.Details()).Methods(http.MethodGet)

	tFlow := templates.Code(template_strings.TransferFlow, cfg.ChainId)

	tFlowBytes, err := json.Marshal(tFlow)
	if err != nil {
		t.Fatal(err)
	}

	validTransferFlow := fmt.Sprintf(`{
		"code":%s,
		"arguments":[{"type":"UFix64","value":"1.0"},{"type":"Address","value":"0xf8d6e0586b0a20c7"}]
	}`, tFlowBytes)

	validHello := `{
		"code":"transaction(greeting: String) { execute { log(greeting.concat(\", World!\")) }}",
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
			name:     "HTTP GET list db empty",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `\[\]`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET list db empty invalid address",
			method:   http.MethodGet,
			url:      "/invalid-address/transactions",
			expected: "not a valid address",
			status:   http.StatusBadRequest,
		},
		{
			name:        "HTTP POST list ok async",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    `"jobId":".+"`,
			status:      http.StatusCreated,
		},
		{
			name:        "HTTP POST list ok sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    `"transactionId":".+"`,
			status:      http.StatusCreated,
			sync:        true,
		},
		{
			name:        "HTTP POST list invalid content-type",
			method:      http.MethodPost,
			contentType: "",
			body:        strings.NewReader(validHello),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    "Unsupported content type",
			status:      http.StatusUnsupportedMediaType,
			sync:        true,
		},
		{
			name:        "HTTP POST list ok sync requires authorizer",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validTransferFlow),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    `"transactionId":".+"`,
			status:      http.StatusCreated,
			sync:        true,
		},
		{
			name:        "HTTP POST list empty body sync",
			method:      http.MethodPost,
			contentType: "application/json",
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    "empty body",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "HTTP POST list invalid body sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader("notvalidobject"),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    "invalid body",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "HTTP POST list invalid code sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(invalidHello),
			url:         "/f8d6e0586b0a20c7/transactions",
			expected:    "Parsing failed",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:        "HTTP POST list invalid address sync",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(validHello),
			url:         "/invalid-address/transactions",
			expected:    "not a valid address",
			status:      http.StatusBadRequest,
			sync:        true,
		},
		{
			name:     "HTTP GET list db not empty",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions",
			expected: `\[.*"transactionId":".+".*\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET details invalid id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/invalid-id",
			expected: "not a valid transaction id",
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP GET details unknown id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/0e4f500d6965c7fc0ff1239525e09eb9dd27c00a511976e353d9f6a44ca22921",
			expected: "transaction not found",
			status:   http.StatusNotFound,
		},
		{
			name:     "HTTP GET details known id",
			method:   http.MethodGet,
			url:      "/f8d6e0586b0a20c7/transactions/<id>",
			expected: `"transactionId":".+"`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET details invalid address",
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
				req.Header.Set(handlers.SYNC_HEADER, "go ahead")
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

func TestScriptsHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer fc.Close()

	service := transactions.NewService(nil, nil, fc, nil)
	h := handlers.NewTransactions(logger, service)

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
			name:   "HTTP POST int 1",
			method: http.MethodPost,
			body: strings.NewReader(`{
				"code":"pub fun main(): Int { return 1 }",
				"arguments":[]
			}`),
			contentType: "application/json",
			expected:    "{\"Value\":1}",
			status:      http.StatusOK,
		},
		{
			name:   "HTTP POST get supply",
			method: http.MethodPost,
			body: strings.NewReader(`{
				"code":"import FlowToken from 0x0ae53cb6e3f42a79\npub fun main(): UFix64 {\nlet supply = FlowToken.totalSupply\nreturn supply\n}",
				"arguments":[]
			}`),
			contentType: "application/json",
			expected:    "1000000000000000000",
			status:      http.StatusOK,
		},
		{
			name:   "HTTP POST get balance",
			method: http.MethodPost,
			body: strings.NewReader(`{
				"code":"import FungibleToken from 0xee82856bf20e2aa6\nimport FlowToken from 0x0ae53cb6e3f42a79\npub fun main(account: Address): UFix64 {\nlet vaultRef = getAccount(account)\n.getCapability(/public/flowTokenBalance)\n.borrow<&FlowToken.Vault{FungibleToken.Balance}>()\n?? panic(\"Could not borrow Balance reference to the Vault\")\nreturn vaultRef.balance\n}",
				"arguments":[{"type":"Address","value":"0xf8d6e0586b0a20c7"}]
			}`),
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
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}
}

func TestTokenServices(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
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
	transactionStore := transactions.NewGormStore(db)

	km := basic.NewKeyManager(keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	accountService := accounts.NewService(accountStore, km, fc, wp)
	transactionService := transactions.NewService(transactionStore, km, fc, wp)
	service := tokens.NewService(transactionService)

	t.Run("account can make a transaction", func(t *testing.T) {
		// Create an account
		_, account, err := accountService.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		// Fund the account from service account
		_, _, err = service.CreateFtWithdrawal(
			context.Background(),
			true,
			"flow-token",
			os.Getenv("ADMIN_ADDRESS"),
			account.Address,
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		_, tx, err := service.CreateFtWithdrawal(
			context.Background(),
			true,
			"flow-token",
			account.Address,
			os.Getenv("ADMIN_ADDRESS"),
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		if flow.HexToID(tx.TransactionId) == flow.EmptyID {
			t.Fatalf("Expected txId not to be empty")
		}
	})

	t.Run("account can not make a transaction without funds", func(t *testing.T) {
		// Create an account
		_, account, err := accountService.Create(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}

		_, tx, err := service.CreateFtWithdrawal(
			context.Background(),
			true,
			"flow-token",
			account.Address,
			os.Getenv("ADMIN_ADDRESS"),
			"1.0",
		)

		if flow.HexToID(tx.TransactionId) == flow.EmptyID {
			t.Fatal("Expected txId not to be empty")
		}

		if err == nil {
			t.Fatal("Expected an error")
		}
	})
}

func TestTokenHandlers(t *testing.T) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
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
	transactionStore := transactions.NewGormStore(db)

	km := basic.NewKeyManager(keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore)
	defer wp.Stop()
	wp.AddWorker(1)

	accountService := accounts.NewService(accountStore, km, fc, wp)
	transactionService := transactions.NewService(transactionStore, km, fc, wp)
	service := tokens.NewService(transactionService)

	h := handlers.NewFungibleTokens(logger, service)

	router := mux.NewRouter()
	router.Handle("/{address}/fungible-tokens/{tokenName}/withdrawals", h.CreateWithdrawal()).Methods(http.MethodPost)

	_, account, err := accountService.Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	createWithdrawalSteps := []httpTestStep{
		{
			name:        "HTTP POST CreateWithdrawal FlowToken valid async",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":"1.0"}`, account.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/flow-token/withdrawals", os.Getenv("ADMIN_ADDRESS")),
			expected:    `"jobId":".+"`,
			status:      http.StatusCreated,
		},
		{
			name:        "HTTP POST CreateWithdrawal FlowToken valid sync",
			sync:        true,
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":"1.0"}`, account.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/flow-token/withdrawals", os.Getenv("ADMIN_ADDRESS")),
			expected:    `"transactionId":".+"`,
			status:      http.StatusCreated,
		},
		{
			name:        "HTTP POST CreateWithdrawal FlowToken invalid recipient",
			method:      http.MethodPost,
			body:        strings.NewReader(`{"recipient":"","amount":"1.0"}`),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/flow-token/withdrawals", os.Getenv("ADMIN_ADDRESS")),
			expected:    "not a valid address",
			status:      http.StatusBadRequest,
		},
		{
			name:        "HTTP POST CreateWithdrawal FlowToken invalid amount",
			method:      http.MethodPost,
			body:        strings.NewReader(fmt.Sprintf(`{"recipient":"%s","amount":""}`, account.Address)),
			contentType: "application/json",
			url:         fmt.Sprintf("/%s/fungible-tokens/flow-token/withdrawals", os.Getenv("ADMIN_ADDRESS")),
			expected:    "missing decimal point",
			status:      http.StatusBadRequest,
		},
	}

	for _, s := range createWithdrawalSteps {
		t.Run(s.name, func(t *testing.T) {
			handleStepRequest(s, router, t)
		})
	}
}
