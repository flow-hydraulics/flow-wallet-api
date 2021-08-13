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
	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/chain_events"
	"github.com/eqlabs/flow-wallet-api/datastore/gorm"
	"github.com/eqlabs/flow-wallet-api/debug"
	"github.com/eqlabs/flow-wallet-api/handlers"
	"github.com/eqlabs/flow-wallet-api/jobs"
	"github.com/eqlabs/flow-wallet-api/keys"
	"github.com/eqlabs/flow-wallet-api/keys/basic"
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/eqlabs/flow-wallet-api/tokens"
	"github.com/eqlabs/flow-wallet-api/transactions"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

const version = "0.5.0"

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

type Config struct {
	Host          string       `env:"HOST"`
	Port          int          `env:"PORT" envDefault:"3000"`
	AccessAPIHost string       `env:"ACCESS_API_HOST,notEmpty"`
	ChainID       flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
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
		fmt.Printf("v%s build on %s from sha1 %s\n", version, buildTime, sha1ver)
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
	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
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
	accountService := accounts.NewService(accountStore, km, fc, wp, transactionService, templateService)
	tokenService := tokens.NewService(tokenStore, km, fc, transactionService, templateService, accountService)

	debugService := debug.Service{
		RepoUrl:   "https://github.com/eqlabs/flow-wallet-api",
		Sha1ver:   sha1ver,
		BuildTime: buildTime,
	}

	// Register a handler for account added events
	accounts.AccountAdded.Register(&tokens.AccountAddedHandler{
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	err = accountService.InitAdminAccount(context.Background(), transactionService)
	if err != nil {
		ls.Fatal(err)
	}

	// HTTP handling

	templateHandler := handlers.NewTemplates(ls, templateService)
	jobsHandler := handlers.NewJobs(ls, jobsService)
	accountHandler := handlers.NewAccounts(ls, accountService)
	transactionHandler := handlers.NewTransactions(ls, transactionService)
	tokenHandler := handlers.NewTokens(ls, tokenService)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Debug
	rv.HandleFunc("/debug", debugService.HandleDebug).Methods(http.MethodGet)

	// Jobs
	rv.Handle("/jobs", jobsHandler.List()).Methods(http.MethodGet)            // list
	rv.Handle("/jobs/{jobId}", jobsHandler.Details()).Methods(http.MethodGet) // details

	// Token templates
	rv.Handle("/tokens", templateHandler.AddToken()).Methods(http.MethodPost)             // create
	rv.Handle("/tokens", templateHandler.ListTokens(nil)).Methods(http.MethodGet)         // list
	rv.Handle("/tokens/{id_or_name}", templateHandler.GetToken()).Methods(http.MethodGet) // details
	rv.Handle("/tokens/{id}", templateHandler.RemoveToken()).Methods(http.MethodDelete)   // delete

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
		tokenType := templates.FT

		// List enabled tokens
		rv.Handle("/fungible-tokens", templateHandler.ListTokens(&tokenType)).Methods(http.MethodGet)

		// Handle "/accounts/{address}/fungible-tokens"
		rft := ra.PathPrefix("/{address}/fungible-tokens").Subrouter()
		rft.Handle("", tokenHandler.AccountTokens(tokenType)).Methods(http.MethodGet)
		rft.Handle("/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
		rft.Handle("/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
		rft.Handle("/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
		rft.Handle("/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)
	} else {
		ls.Println("fungible tokens disabled")
	}

	// Non-Fungible tokens
	if !disableNft {
		tokenType := templates.NFT

		// List enabled tokens
		rv.Handle("/non-fungible-tokens", templateHandler.ListTokens(&tokenType)).Methods(http.MethodGet)

		rnft := ra.PathPrefix("/{address}/non-fungible-tokens").Subrouter()
		rnft.Handle("", tokenHandler.AccountTokens(tokenType)).Methods(http.MethodGet)
		rnft.Handle("/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
		rnft.Handle("/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
		rnft.Handle("/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
		rnft.Handle("/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
		rnft.Handle("/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
		rnft.Handle("/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
		rnft.Handle("/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)
	} else {
		ls.Println("non-fungible tokens disabled")
	}

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
		ls.Println("Starting chain event listener..")

		store := chain_events.NewGormStore(db)
		maxDiff := uint64(100)       // maximum number of blocks to check each iteration, TODO: make this configurable
		interval := 10 * time.Second // TODO: make this configurable
		getTypes := func() []string {
			// Get all enabled tokens
			tt, err := templateService.ListTokens(nil)
			if err != nil {
				panic(err)
			}

			types := make([]string, len(*tt))

			// Listen for enabled tokens deposit events
			for i, token := range *tt {
				types[i] = templates.DepositEventTypeFromToken(token)
			}

			return types
		}

		listener := chain_events.NewListener(fc, store, maxDiff, interval, getTypes)

		defer func() {
			ls.Println("Stopping event listener..")
			listener.Stop()
		}()

		// Register a handler for chain events
		chain_events.Event.Register(&tokens.ChainEventHandler{
			AccountService:  accountService,
			ChainListener:   listener,
			TemplateService: templateService,
			TokenService:    tokenService,
		})

		listener.Start()
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
