package main

import (
	"context"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/account"
	"github.com/eqlabs/flow-wallet-service/data/gorm"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys/simple"
	"github.com/eqlabs/flow-wallet-service/tokens"
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

func TestMain(m *testing.M) {
	godotenv.Load(".env.test")

	os.Setenv("DB_DSN", testDbDSN)
	os.Setenv("DB_TYPE", testDbType)

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	logger = log.New(io.Discard, "", log.LstdFlags)

	exitcode := m.Run()

	os.Remove(testDbDSN)
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

	db, err := gorm.NewStore(logger)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	km, err := simple.NewKeyManager(logger, db, fc)
	if err != nil {
		t.Fatal(err)
	}

	wp := jobs.NewWorkerPool(logger, db)
	defer wp.Stop()
	wp.AddWorker(1)

	service := account.NewService(logger, db, km, fc, wp)

	t.Run("sync create", func(t *testing.T) {
		account, err := service.Create(context.Background())
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

	t.Run("async create", func(t *testing.T) {
		job, err := service.CreateAsync()
		if err != nil {
			t.Fatal(err)
		}

		if job.Status != jobs.STATUS_ACCEPTED && job.Status != jobs.STATUS_COMPLETE {
			t.Errorf("expected job status to be %s or %s but got %s",
				jobs.STATUS_ACCEPTED, jobs.STATUS_COMPLETE, job.Status)
		}

		for job.Status == jobs.STATUS_ACCEPTED {
			time.Sleep(10 * time.Millisecond)
		}

		if job.Status != jobs.STATUS_COMPLETE {
			t.Errorf("expected job status to be %s got %s", jobs.STATUS_COMPLETE, job.Status)
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

	// TODO: these will need to be async as well
	t.Run("account can make a transaction", func(t *testing.T) {
		account, err := service.Create(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		// Fund the account from service account
		txId, err := tokens.TransferFlow(
			km,
			fc,
			flow.HexToAddress(account.Address),
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
			km,
			fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(account.Address),
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

	// TODO: these will need to be async as well
	t.Run("account can not make a transaction without funds", func(t *testing.T) {
		account, err := service.Create(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		txId, err := tokens.TransferFlow(
			km,
			fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(account.Address),
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
