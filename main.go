package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/handlers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store/memory"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store/simple"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type config struct {
	Host                   string `env:"HOST" envDefault:""`
	Port                   int    `env:"PORT" envDefault:"3000"`
	DefaultKeyManager      string `env:"DEFAULT_KEY_MANAGER" envDefault:"local"`
	FlowGateway            string `env:"FLOW_GATEWAY" envDefault:"localhost:3569"`
	EncryptionKey          string `env:"ENCRYPTION_KEY" envDefault:""`
	ServiceAccountAddress  string `env:"SERVICE_ACC_ADDRESS,required"`
	ServiceAccountKeyIndex int    `env:"SERVICE_ACC_KEY_INDEX" envDefault:"0"`
	ServiceAccountKeyType  string `env:"SERVICE_ACC_KEY_TYPE" envDefault:"local"`
	ServiceAccountKeyValue string `env:"SERVICE_ACC_KEY_VALUE,required"`
}

func main() {
	godotenv.Load(".env")

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	var (
		disable_account_mgmt bool
		disable_raw_tx       bool
		disable_ft           bool
		// disable_nft          bool
	)

	flag.BoolVar(&disable_account_mgmt, "disable-account-mgmt", false, "disable account management")
	flag.BoolVar(&disable_raw_tx, "disable-raw-tx", false, "disable sending raw transactions for an account")
	flag.BoolVar(&disable_ft, "disable-ft", false, "disable fungible token functionality")
	flag.Parse()

	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// TODO: WithInsecure()?
	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	db, err := memory.NewDataStore()
	if err != nil {
		log.Fatal(err)
	}

	ks, err := simple.NewKeyStore(
		db,
		store.AccountKey{
			AccountAddress: flow.HexToAddress(cfg.ServiceAccountAddress),
			Index:          cfg.ServiceAccountKeyIndex,
			Type:           cfg.ServiceAccountKeyType,
			Value:          cfg.ServiceAccountKeyValue,
		},
		cfg.DefaultKeyManager,
		cfg.EncryptionKey,
	)
	if err != nil {
		log.Fatal(err)
	}

	accounts := handlers.NewAccounts(l, fc, db, ks)
	transactions := handlers.NewTransactions(l, fc, db, ks)
	fungibleTokens := handlers.NewFungibleTokens(l, fc, db, ks)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Handle "/accounts"
	ra := rv.PathPrefix("/accounts").Subrouter()

	// Account list
	ra.HandleFunc("", accounts.List).Methods("GET")

	// Account create
	if !disable_account_mgmt {
		ra.HandleFunc("", accounts.Create).Methods("POST")
	}

	// Account details
	ra.HandleFunc("/{address}", accounts.Details).Methods("GET")

	// Account delete
	if !disable_account_mgmt {
		ra.HandleFunc("/{address}", accounts.Delete).Methods("DELETE")
	}

	// Account send raw transaction
	if !disable_raw_tx {
		ra.HandleFunc("/{address}/transactions", transactions.SendTransaction).Methods("POST")
	}

	// Fungible tokens
	if !disable_ft {
		// Handle "/accounts/{address}/fungible-tokens"
		rft := ra.PathPrefix("/{address}/fungible-tokens").Subrouter()
		rft.HandleFunc("/{tokenName}", fungibleTokens.Details).Methods("GET")
		rft.HandleFunc("/{tokenName}", fungibleTokens.Init).Methods("POST")
		rft.HandleFunc("/{tokenName}/withdrawals", fungibleTokens.ListWithdrawals).Methods("GET")
		rft.HandleFunc("/{tokenName}/withdrawals", fungibleTokens.CreateWithdrawal).Methods("POST")
		rft.HandleFunc("/{tokenName}/withdrawals/{transactionId}", fungibleTokens.WithdrawalDetails).Methods("GET")
	}

	// TODO: nfts

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		l.Println("Server running")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// Trap interupt or sigterm and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	l.Printf("Got signal: %s. Shutting down..\n", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
