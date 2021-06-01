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
	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/eqlabs/flow-wallet-service/datastore/gorm"
	"github.com/eqlabs/flow-wallet-service/debug"
	"github.com/eqlabs/flow-wallet-service/handlers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/keys/basic"
	"github.com/eqlabs/flow-wallet-service/tokens"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

type Config struct {
	Host          string `env:"HOST"`
	Port          int    `env:"PORT" envDefault:"3000"`
	AccessApiHost string `env:"ACCESS_API_HOST,required"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("WARNING: Could not load environment variables from file; ", err)
	}

	var (
		flgVersion   bool
		flgRunServer bool
	)

	flag.BoolVar(&flgVersion, "version", false, "if true, print version and exit")
	flag.BoolVar(&flgRunServer, "server", true, "run the server")
	flag.Parse()

	if flgVersion {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		os.Exit(0)
	}

	if flgRunServer {
		runServer()
		os.Exit(0)
	}
}

func runServer() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	var (
		flgDisableRawTx bool
		flgDisableFt    bool
		// flgDisableNft    bool
	)

	flag.BoolVar(&flgDisableRawTx, "disable-raw-tx", false, "disable raw transactions api")
	flag.BoolVar(&flgDisableFt, "disable-ft", false, "disable fungible token api")
	// flag.BoolVar(&flgDisableNft, "disable-nft", false, "disable non-fungible token functionality")
	flag.Parse()

	// Application wide logger
	ls := log.New(os.Stdout, "[SERVER] ", log.LstdFlags|log.Lshortfile)
	lj := log.New(os.Stdout, "[JOBS] ", log.LstdFlags|log.Lshortfile)

	// Flow client
	// TODO: WithInsecure()?
	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Println("Closing Flow Client..")
		err := fc.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Database
	db, err := gorm.New()
	if err != nil {
		log.Fatal(err)
	}
	defer gorm.Close(db)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)

	// Create a worker pool
	wp := jobs.NewWorkerPool(lj, jobStore)
	// TODO: make this configurable
	wp.AddWorker(100) // Add a worker with capacity of 100

	// Key manager
	km := basic.NewKeyManager(keyStore, fc)

	// Services
	jobsService := jobs.NewService(jobStore)
	accountService := accounts.NewService(accountStore, km, fc, wp)
	transactionService := transactions.NewService(transactionStore, km, fc, wp)
	tokenService := tokens.NewService(transactionService)

	debugService := debug.Service{
		RepoUrl:   "https://github.com/eqlabs/flow-wallet-service",
		Sha1ver:   sha1ver,
		BuildTime: buildTime,
	}

	// HTTP handling

	jobsHandler := handlers.NewJobs(ls, jobsService)
	accountsHandler := handlers.NewAccounts(ls, accountService)
	transactions := handlers.NewTransactions(ls, transactionService)
	fungibleTokens := handlers.NewFungibleTokens(ls, tokenService)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Debug
	rv.HandleFunc("/debug", debugService.HandleDebug).Methods(http.MethodGet) // details

	// Jobs
	rv.Handle("/jobs", jobsHandler.List()).Methods(http.MethodGet)            // details
	rv.Handle("/jobs/{jobId}", jobsHandler.Details()).Methods(http.MethodGet) // details

	// Account
	ra := rv.PathPrefix("/accounts").Subrouter()
	ra.Handle("", accountsHandler.List()).Methods(http.MethodGet)              // list
	ra.Handle("", accountsHandler.Create()).Methods(http.MethodPost)           // create
	ra.Handle("/{address}", accountsHandler.Details()).Methods(http.MethodGet) // details

	// Account raw transactions
	if !flgDisableRawTx {
		rt := rv.PathPrefix("/accounts/{address}/transactions").Subrouter()
		rt.Handle("", transactions.List()).Methods(http.MethodGet)                    // list
		rt.Handle("", transactions.Create()).Methods(http.MethodPost)                 // create
		rt.Handle("/{transactionId}", transactions.Details()).Methods(http.MethodGet) // details
	} else {
		ls.Println("raw transactions disabled")
	}

	// Scripts
	rv.Handle("/scripts", transactions.ExecuteScript()).Methods(http.MethodPost) // create

	// Fungible tokens
	if !flgDisableFt {
		// Handle "/accounts/{address}/fungible-tokens"
		rft := ra.PathPrefix("/{address}/fungible-tokens").Subrouter()
		// rft.Handle("/{tokenName}", fungibleTokens.Details).Methods(http.MethodGet)
		// rft.Handle("/{tokenName}/withdrawals", fungibleTokens.ListWithdrawals).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/withdrawals", fungibleTokens.CreateWithdrawal()).Methods(http.MethodPost)
		// rft.Handle("/{tokenName}/withdrawals/{transactionId}", fungibleTokens.WithdrawalDetails).Methods(http.MethodGet)
	}

	// TODO: nfts

	// Define middleware
	h := handlers.UseCors(r)
	h = handlers.UseLogging(os.Stdout, h)
	h = handlers.UseCompress(h)

	// Server boilerplate
	srv := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		ls.Printf("Listening on port %d\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			ls.Println(err)
		}
	}()

	// Trap interupt or sigterm and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	ls.Printf("Got signal: %s. Shutting down..\n", sig)

	// Stop the worker pool, waits
	wp.Stop()

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Error in server shutdown; ", err)
	}
}
