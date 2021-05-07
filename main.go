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
	"github.com/eqlabs/flow-wallet-service/account"
	"github.com/eqlabs/flow-wallet-service/data/gorm"
	"github.com/eqlabs/flow-wallet-service/handlers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys/simple"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type Config struct {
	Host        string `env:"HOST"`
	Port        int    `env:"PORT" envDefault:"3000"`
	FlowGateway string `env:"FLOW_GATEWAY,required"`
}

func main() {
	godotenv.Load(".env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	var (
		disable_raw_tx bool
		disable_ft     bool
		disable_nft    bool
	)

	flag.BoolVar(&disable_raw_tx, "disable-raw-tx", false, "disable sending raw transactions for an account")
	flag.BoolVar(&disable_ft, "disable-ft", false, "disable fungible token functionality")
	flag.BoolVar(&disable_nft, "disable-nft", false, "disable non-fungible token functionality")
	flag.Parse()

	// Application wide logger
	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// Flow client
	// TODO: WithInsecure()?
	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	// Database
	db, err := gorm.NewStore(l)
	if err != nil {
		log.Fatal(err)
	}

	// Create a worker pool
	wp := jobs.NewWorkerPool(l, db)
	wp.AddWorker(100) // Add a worker with capacity of 100

	// Key manager
	km, err := simple.NewKeyManager(l, db, fc)
	if err != nil {
		log.Fatal(err)
	}

	// Services
	jobsService := jobs.NewService(l, db)
	accountService := account.NewService(l, db, km, fc, wp)

	// HTTP handling

	jobsHandler := handlers.NewJobs(l, jobsService)
	accountsHandler := handlers.NewAccounts(l, accountService)
	// transactions := handlers.NewTransactions(l, fc, db, km)
	// fungibleTokens := handlers.NewFungibleTokens(l, fc, db, km)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Jobs
	rv.HandleFunc("/status/{jobId}", jobsHandler.Details).Methods(http.MethodGet) // details

	// Account
	ra := rv.PathPrefix("/accounts").Subrouter()
	ra.HandleFunc("", accountsHandler.List).Methods(http.MethodGet)              // list
	ra.HandleFunc("", accountsHandler.Create).Methods(http.MethodPost)           // create
	ra.HandleFunc("/{address}", accountsHandler.Details).Methods(http.MethodGet) // details

	// // Account raw transactions
	// if !disable_raw_tx {
	// 	rt := rv.PathPrefix("/accounts/{address}/transactions").Subrouter()
	// 	rt.HandleFunc("", transactions.List).Methods(http.MethodGet)                    // list
	// 	rt.HandleFunc("", transactions.Create).Methods(http.MethodPost)                 // create
	// 	rt.HandleFunc("/{transactionId}", transactions.Details).Methods(http.MethodGet) // details
	// }

	// // Fungible tokens
	// if !disable_ft {
	// 	// Handle "/accounts/{address}/fungible-tokens"
	// 	rft := ra.PathPrefix("/{address}/fungible-tokens").Subrouter()
	// 	rft.HandleFunc("/{tokenName}", fungibleTokens.Details).Methods(http.MethodGet)
	// 	rft.HandleFunc("/{tokenName}", fungibleTokens.Init).Methods(http.MethodPost)
	// 	rft.HandleFunc("/{tokenName}/withdrawals", fungibleTokens.ListWithdrawals).Methods(http.MethodGet)
	// 	rft.HandleFunc("/{tokenName}/withdrawals", fungibleTokens.CreateWithdrawal).Methods(http.MethodPost)
	// 	rft.HandleFunc("/{tokenName}/withdrawals/{transactionId}", fungibleTokens.WithdrawalDetails).Methods(http.MethodGet)
	// }

	// TODO: nfts

	// Server boilerplate

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

	// Stop the worker pool, waits
	wp.Stop()

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
