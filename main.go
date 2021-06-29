package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/eqlabs/flow-wallet-service/datastore/gorm"
	"github.com/eqlabs/flow-wallet-service/debug"
	"github.com/eqlabs/flow-wallet-service/events"
	"github.com/eqlabs/flow-wallet-service/handlers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/keys/basic"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/tokens"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

type Config struct {
	Host          string       `env:"HOST"`
	Port          int          `env:"PORT" envDefault:"3000"`
	AccessApiHost string       `env:"ACCESS_API_HOST,required"`
	ChainId       flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Could not load environment variables from file.\n%s\nIf running inside a docker container this can be ignored.\n\n", err)
	}

	var (
		flgVersion            bool
		flgDisableRawTx       bool
		flgDisableFt          bool
		flgDisableNft         bool
		flgDisableChainEvents bool
	)

	flag.BoolVar(&flgVersion, "version", false, "if true, print version and exit")
	flag.BoolVar(&flgDisableRawTx, "disable-raw-tx", false, "disable raw transactions api")
	flag.BoolVar(&flgDisableFt, "disable-ft", false, "disable fungible token api")
	flag.BoolVar(&flgDisableNft, "disable-nft", false, "disable non-fungible token functionality")
	flag.BoolVar(&flgDisableChainEvents, "disable-events", false, "disable chain event listener")

	flag.Parse()

	if flgVersion {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		os.Exit(0)
	}

	runServer(
		flgDisableRawTx,
		flgDisableFt,
		flgDisableNft,
		flgDisableChainEvents,
	)

	os.Exit(0)
}

func runServer(disableRawTx, disableFt, disableNft, disableChainEvents bool) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	// Application wide logger
	ls := log.New(os.Stdout, "[SERVER] ", log.LstdFlags|log.Lshortfile)
	lj := log.New(os.Stdout, "[JOBS] ", log.LstdFlags|log.Lshortfile)

	// Flow client
	// TODO: WithInsecure()?
	fc, err := client.New(cfg.AccessApiHost, grpc.WithInsecure())
	if err != nil {
		ls.Fatal(err)
	}
	defer func() {
		ls.Println("Closing Flow Client..")
		if err := fc.Close(); err != nil {
			ls.Fatal(err)
		}
	}()

	// Database
	db, err := gorm.New()
	if err != nil {
		ls.Fatal(err)
	}
	defer gorm.Close(db)

	templateStore := templates.NewGormStore(db)
	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)

	// Create a worker pool
	wp := jobs.NewWorkerPool(lj, jobStore)
	defer func() {
		ls.Println("Stopping worker pool..")
		wp.Stop()
	}()

	// TODO: make this configurable
	wp.AddWorker(100) // Add a worker with capacity of 100

	// Key manager
	km := basic.NewKeyManager(keyStore, fc)

	// Services
	templateService := templates.NewService(templateStore)
	jobsService := jobs.NewService(jobStore)
	transactionService := transactions.NewService(transactionStore, km, fc, wp)
	accountService := accounts.NewService(accountStore, km, fc, wp, transactionService)
	tokenService := tokens.NewService(tokenStore, km, fc, transactionService)

	debugService := debug.Service{
		RepoUrl:   "https://github.com/eqlabs/flow-wallet-service",
		Sha1ver:   sha1ver,
		BuildTime: buildTime,
	}

	accountService.InitAdminAccount()

	// HTTP handling

	templateHandler := handlers.NewTemplates(ls, templateService)
	jobsHandler := handlers.NewJobs(ls, jobsService)
	accountHandler := handlers.NewAccounts(ls, accountService)
	transactionHandler := handlers.NewTransactions(ls, transactionService)
	fungibleTokenHandler := handlers.NewFungibleTokens(ls, tokenService)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Debug
	rv.HandleFunc("/debug", debugService.HandleDebug).Methods(http.MethodGet)

	// Jobs
	rv.Handle("/jobs", jobsHandler.List()).Methods(http.MethodGet)            // list
	rv.Handle("/jobs/{jobId}", jobsHandler.Details()).Methods(http.MethodGet) // details

	// Token templates
	rv.Handle("/tokens", templateHandler.AddToken()).Methods(http.MethodPost)           // create
	rv.Handle("/tokens", templateHandler.ListTokens()).Methods(http.MethodGet)          // list
	rv.Handle("/tokens/{id}", templateHandler.GetToken()).Methods(http.MethodGet)       // details
	rv.Handle("/tokens/{id}", templateHandler.RemoveToken()).Methods(http.MethodDelete) // delete

	// Account
	ra := rv.PathPrefix("/accounts").Subrouter()
	ra.Handle("", accountHandler.List()).Methods(http.MethodGet)              // list
	ra.Handle("", accountHandler.Create()).Methods(http.MethodPost)           // create
	ra.Handle("/{address}", accountHandler.Details()).Methods(http.MethodGet) // details

	// Account raw transactions
	if !disableRawTx {
		rt := rv.PathPrefix("/accounts/{address}/transactions").Subrouter()
		rt.Handle("", transactionHandler.List()).Methods(http.MethodGet)                    // list
		rt.Handle("", transactionHandler.Create()).Methods(http.MethodPost)                 // create
		rt.Handle("/{transactionId}", transactionHandler.Details()).Methods(http.MethodGet) // details
	} else {
		ls.Println("raw transactions disabled")
	}

	// Scripts
	rv.Handle("/scripts", transactionHandler.ExecuteScript()).Methods(http.MethodPost) // create

	// Fungible tokens
	if !disableFt {
		// List enabled tokens
		rv.Handle("/fungible-tokens", fungibleTokenHandler.List()).Methods(http.MethodGet)

		// Handle "/accounts/{address}/fungible-tokens"
		rft := ra.PathPrefix("/{address}/fungible-tokens").Subrouter()
		rft.Handle("", accountHandler.AccountFungibleTokens()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}", fungibleTokenHandler.Details()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}", accountHandler.SetupFungibleToken()).Methods(http.MethodPost)
		rft.Handle("/{tokenName}/withdrawals", fungibleTokenHandler.ListFtWithdrawals()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/withdrawals", fungibleTokenHandler.CreateFtWithdrawal()).Methods(http.MethodPost)
		rft.Handle("/{tokenName}/withdrawals/{transactionId}", fungibleTokenHandler.GetFtWithdrawal()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/deposits", fungibleTokenHandler.ListFtDeposits()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/deposits/{transactionId}", fungibleTokenHandler.GetFtDeposit()).Methods(http.MethodGet)
	} else {
		ls.Println("fungible tokens disabled")
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
		ls.Printf("Server listening on port %d\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			ls.Println(err)
		}
	}()

	// Chain event listener
	if !disableChainEvents {
		ls.Println("Starting event listener..")

		eventStore := events.NewGormStore(db)

		l := events.
			NewListener(fc, eventStore, 100).
			Start()

		defer func() {
			ls.Println("Stopping event listener..")
			l.Stop()
		}()

		// Listen for enabled tokens deposit events
		for _, t := range templates.EnabledTokens() {
			l.ListenTokenEvent(t, events.TokensDeposited)
		}

		go func() {
			for ee := range l.Events {
				for _, e := range ee {
					ss := strings.Split(e.Type, ".")
					if ss[len(ss)-1] == events.TokensDeposited {
						t, err := templates.TokenFromEvent(e, cfg.ChainId)
						if err != nil {
							continue
						}
						if !t.IsEnabled() {
							continue
						}

						// Check if recipient is in database
						a, err := accountService.Details(e.Value.Fields[1].String())
						if err != nil {
							continue
						}

						if err = tokenService.RegisterFtDeposit(t, e.TransactionID.Hex(), e.Value.Fields[0].String(), a.Address); err != nil {
							ls.Printf("error while registering a deposit: %s\n", err)
						}
					}
				}
			}
		}()
	}

	// Trap interupt or sigterm and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	ls.Printf("Got signal: %s. Shutting down..\n", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		ls.Fatal("Error in server shutdown; ", err)
	}
}
