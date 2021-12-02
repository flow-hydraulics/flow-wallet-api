package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/chain_events"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/handlers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/gorilla/mux"
	"github.com/onflow/flow-go-sdk/client"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
)

const version = "0.8.0"

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func main() {
	var (
		printVersion bool
		envFilePath  string
	)

	// If we should just print the version number and exit
	flag.BoolVar(&printVersion, "version", false, "if true, print version and exit")

	// Allow configuration of envfile path
	// If not set, ParseConfig will not try to load variables to environment from a file
	flag.StringVar(&envFilePath, "envfile", "", "envfile path")

	flag.Parse()

	if printVersion {
		fmt.Printf("v%s build on %s from sha1 %s\n", version, buildTime, sha1ver)
		os.Exit(0)
	}

	opts := &configs.Options{EnvFilePath: envFilePath, Version: version}
	cfg, err := configs.ParseConfig(opts)
	if err != nil {
		panic(err)
	}

	runServer(cfg)

	os.Exit(0)
}

func runServer(cfg *configs.Config) {
	cfg.Logger.Info("Starting server")

	// Flow client
	// TODO: WithInsecure()?
	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		cfg.Logger.Fatal(err)
	}
	defer func() {
		cfg.Logger.Info("Closing Flow Client")
		if err := fc.Close(); err != nil {
			cfg.Logger.Fatal(err)
		}
	}()

	// Database
	db, err := gorm.New(cfg)
	if err != nil {
		cfg.Logger.Fatal(err)
	}
	defer gorm.Close(db)

	systemStore := system.NewGormStore(db)
	templateStore := templates.NewGormStore(db)
	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)

	systemService := system.NewService(systemStore)

	// Create a worker pool
	wp := jobs.NewWorkerPool(
		cfg.Logger,
		jobStore,
		cfg.WorkerQueueCapacity,
		cfg.WorkerCount,
		jobs.WithJobStatusWebhook(cfg.JobStatusWebhookUrl, cfg.JobStatusWebhookTimeout),
		jobs.WithSystemService(systemService),
	)
	defer func() {
		cfg.Logger.Info("Stopping worker pool")
		wp.Stop()
	}()

	txRatelimiter := ratelimit.New(cfg.TransactionMaxSendRate, ratelimit.WithoutSlack)

	// Key manager
	km := basic.NewKeyManager(cfg, keyStore, fc)

	// Services
	templateService := templates.NewService(cfg, templateStore)
	jobsService := jobs.NewService(jobStore)
	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp, transactions.WithTxRatelimiter(txRatelimiter))
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService, accounts.WithTxRatelimiter(txRatelimiter))
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, transactionService, templateService, accountService)

	// Register a handler for account added events
	accounts.AccountAdded.Register(&tokens.AccountAddedHandler{
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	err = accountService.InitAdminAccount(context.Background())
	if err != nil {
		cfg.Logger.Fatal(err)
	}

	// HTTP handling
	systemHandler := handlers.NewSystem(cfg.Logger, systemService)
	templateHandler := handlers.NewTemplates(cfg.Logger, templateService)
	jobsHandler := handlers.NewJobs(cfg.Logger, jobsService)
	accountHandler := handlers.NewAccounts(cfg.Logger, accountService)
	transactionHandler := handlers.NewTransactions(cfg.Logger, transactionService)
	tokenHandler := handlers.NewTokens(cfg.Logger, tokenService)

	r := mux.NewRouter()

	// Catch the api version
	rv := r.PathPrefix("/{apiVersion}").Subrouter()

	// Debug
	rv.Handle("/debug", handlers.Debug("https://github.com/flow-hydraulics/flow-wallet-api", sha1ver, buildTime)).Methods(http.MethodGet)

	// Health
	rv.HandleFunc("/health/ready", handlers.HandleHealthReady).Methods(http.MethodGet)
	rv.Handle("/health/liveness", handlers.Liveness(func() (interface{}, error) {
		return wp.Status()
	})).Methods(http.MethodGet)

	// System
	rv.Handle("/system/settings", systemHandler.GetSettings()).Methods(http.MethodGet)
	rv.Handle("/system/settings", systemHandler.SetSettings()).Methods(http.MethodPost)

	// Jobs
	rv.Handle("/jobs", jobsHandler.List()).Methods(http.MethodGet)            // list
	rv.Handle("/jobs/{jobId}", jobsHandler.Details()).Methods(http.MethodGet) // details

	// Token templates
	rv.Handle("/tokens", templateHandler.ListTokens(templates.NotSpecified)).Methods(http.MethodGet) // list
	rv.Handle("/tokens", templateHandler.AddToken()).Methods(http.MethodPost)                        // create
	rv.Handle("/tokens/{id_or_name}", templateHandler.GetToken()).Methods(http.MethodGet)            // details
	rv.Handle("/tokens/{id}", templateHandler.RemoveToken()).Methods(http.MethodDelete)              // delete

	// List enabled tokens by type
	rv.Handle("/fungible-tokens", templateHandler.ListTokens(templates.FT)).Methods(http.MethodGet)      // list
	rv.Handle("/non-fungible-tokens", templateHandler.ListTokens(templates.NFT)).Methods(http.MethodGet) // list

	// Transactions
	rv.Handle("/transactions", transactionHandler.List()).Methods(http.MethodGet)                    // list
	rv.Handle("/transactions/{transactionId}", transactionHandler.Details()).Methods(http.MethodGet) // details

	// Account
	rv.Handle("/accounts", accountHandler.List()).Methods(http.MethodGet)              // list
	rv.Handle("/accounts", accountHandler.Create()).Methods(http.MethodPost)           // create
	rv.Handle("/accounts/{address}", accountHandler.Details()).Methods(http.MethodGet) // details

	// Account raw transactions
	if !cfg.DisableRawTransactions {
		rv.Handle("/accounts/{address}/sign", transactionHandler.Sign()).Methods(http.MethodPost)                           // sign
		rv.Handle("/accounts/{address}/transactions", transactionHandler.List()).Methods(http.MethodGet)                    // list
		rv.Handle("/accounts/{address}/transactions", transactionHandler.Create()).Methods(http.MethodPost)                 // create
		rv.Handle("/accounts/{address}/transactions/{transactionId}", transactionHandler.Details()).Methods(http.MethodGet) // details
	} else {
		cfg.Logger.Info("raw transactions disabled")
	}

	// Non-custodial watchlist accounts
	rv.Handle("/watchlist/accounts", accountHandler.AddNonCustodialAccount()).Methods(http.MethodPost)                // add
	rv.Handle("/watchlist/accounts/{address}", accountHandler.DeleteNonCustodialAccount()).Methods(http.MethodDelete) // delete

	// Scripts
	rv.Handle("/scripts", transactionHandler.ExecuteScript()).Methods(http.MethodPost) // create

	// Fungible tokens
	if !cfg.DisableFungibleTokens {
		rv.Handle("/accounts/{address}/fungible-tokens", tokenHandler.AccountTokens(templates.FT)).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/fungible-tokens/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)
	} else {
		cfg.Logger.Info("fungible tokens disabled")
	}

	// Non-Fungible tokens
	if !cfg.DisableNonFungibleTokens {
		rv.Handle("/accounts/{address}/non-fungible-tokens", tokenHandler.AccountTokens(templates.NFT)).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}", tokenHandler.Details()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}", tokenHandler.Setup()).Methods(http.MethodPost)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals", tokenHandler.ListWithdrawals()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals", tokenHandler.CreateWithdrawal()).Methods(http.MethodPost)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals/{transactionId}", tokenHandler.GetWithdrawal()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}/deposits", tokenHandler.ListDeposits()).Methods(http.MethodGet)
		rv.Handle("/accounts/{address}/non-fungible-tokens/{tokenName}/deposits/{transactionId}", tokenHandler.GetDeposit()).Methods(http.MethodGet)
	} else {
		cfg.Logger.Info("non-fungible tokens disabled")
	}

	// Define middleware
	h := http.TimeoutHandler(r, cfg.ServerRequestTimeout, "request timed out")
	h = handlers.UseCors(h)
	h = handlers.UseLogging(cfg.Logger, h)
	h = handlers.UseCompress(h)

	// Server boilerplate
	srv := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		WriteTimeout: 0, // Disabled, set cfg.ServerRequestTimeout instead
		ReadTimeout:  0, // Disabled, set cfg.ServerRequestTimeout instead
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		cfg.Logger.
			WithFields(log.Fields{
				"host": cfg.Host,
				"port": cfg.Port,
			}).
			Info("Server listening")
		if err := srv.ListenAndServe(); err != nil {
			cfg.Logger.Fatal(err)
		}
	}()

	// Chain event listener
	if !cfg.DisableChainEvents {
		store := chain_events.NewGormStore(db)
		getTypes := func() ([]string, error) {
			// Get all enabled tokens
			tt, err := templateService.ListTokens(templates.NotSpecified)
			if err != nil {
				return nil, err
			}

			token_count := len(*tt)
			event_types := make([]string, token_count)

			// Listen for enabled tokens deposit events
			for i, token := range *tt {
				event_types[i] = templates.DepositEventTypeFromToken(token)
			}

			return event_types, nil
		}

		listener := chain_events.NewListener(
			cfg.Logger, fc, store, getTypes,
			cfg.ChainListenerMaxBlocks,
			cfg.ChainListenerInterval,
			cfg.ChainListenerStartingHeight,
			chain_events.WithSystemService(systemService),
		)

		defer listener.Stop()

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

	cfg.Logger.Infof("Got signal: %s. Shutting down..", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		cfg.Logger.Fatalf("Error in server shutdown: %s", err)
	}
}
