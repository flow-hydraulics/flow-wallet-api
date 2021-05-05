package account

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/data/gorm"
	"github.com/eqlabs/flow-wallet-service/keys/simple"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type testConfig struct {
	FlowGateway string `env:"FLOW_GATEWAY,required"`
}

const testDbDSN = "test.db"
const testDbType = "sqlite"

// TestServiceSetup is used to spin up an account service for testing.
func TestServiceSetup(l *log.Logger) (result *Service, err error) {
	godotenv.Load("../.env.test")

	os.Setenv("DB_DSN", testDbDSN)
	os.Setenv("DB_TYPE", testDbType)

	var cfg testConfig
	if err = env.Parse(&cfg); err != nil {
		return
	}

	// Flow client
	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		return
	}

	// Database
	db, err := gorm.NewStore(l)
	if err != nil {
		return
	}

	// Key manager
	km, err := simple.NewKeyManager(l, db, fc)
	if err != nil {
		return
	}

	result = NewService(l, db, km, fc)

	return
}

// TestServiceTearDown is used to clean up after account service testing.
// It currently only deletes the test database files.
func TestServiceTearDown() {
	os.Remove(testDbDSN)
}
