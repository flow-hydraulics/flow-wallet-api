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
	"github.com/flow-hydraulics/flow-wallet-api/ops"
	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	access "github.com/onflow/flow-go-sdk/access/grpc"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const version = "0.9.0"

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func main() {
	var (
		printVersion bool
		envFilePath  string // LEGACY: now used to check if user still is using envFilePath
	)

	// If we should just print the version number and exit
	flag.BoolVar(&printVersion, "version", false, "if true, print version and exit")
	flag.StringVar(&envFilePath, "envfile", "", "deprecated")
	flag.Parse()

	if envFilePath != "" {
		panic("'-envfile' is no longer supported, see readme")
	}

	if printVersion {
		fmt.Printf("v%s build on %s from sha1 %s\n", version, buildTime, sha1ver)
		os.Exit(0)
	}

	cfg, err := configs.Parse()
	if err != nil {
		panic(err)
	}

	runServer(cfg)

	os.Exit(0)
}

func runServer(cfg *configs.Config) {
	configs.ConfigureLogger(cfg.LogLevel)

	log.Info("Starting server")

	// Flow client
	// TODO: WithInsecure()?
	fc, err := access.NewClient(
		cfg.AccessAPIHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.GrpcMaxCallRecvMsgSize)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fc.Close(); err != nil {
			log.Warn(err)
		}
		log.Info("Closed Flow Client")
	}()

	// Database
	db, err := gorm.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer gorm.Close(db)

	systemService := system.NewService(
		system.NewGormStore(db),
		system.WithPauseDuration(cfg.PauseDuration),
	)

	// Create a worker pool
	wp := jobs.NewWorkerPool(
		jobs.NewGormStore(db),
		cfg.WorkerQueueCapacity,
		cfg.WorkerCount,
		jobs.WithJobStatusWebhook(cfg.JobStatusWebhookUrl, cfg.JobStatusWebhookTimeout),
		jobs.WithSystemService(systemService),
		jobs.WithMaxJobErrorCount(cfg.MaxJobErrorCount),
		jobs.WithDbJobPollInterval(cfg.DBJobPollInterval),
		jobs.WithAcceptedGracePeriod(cfg.AcceptedGracePeriod),
		jobs.WithReSchedulableGracePeriod(cfg.ReSchedulableGracePeriod),
	)

	defer func() {
		wp.Stop(true)
		log.Info("Stopped workerpool")
	}()

	txRatelimiter := ratelimit.New(cfg.TransactionMaxSendRate, ratelimit.WithoutSlack)

	// Key manager
	km := basic.NewKeyManager(cfg, keys.NewGormStore(db), fc)

	// Services
	templateService, err := templates.NewService(cfg, templates.NewGormStore(db))
	if err != nil {
		log.Fatal(err)
	}
	jobsService := jobs.NewService(jobs.NewGormStore(db))
	transactionService := transactions.NewService(cfg, transactions.NewGormStore(db), km, fc, wp, transactions.WithTxRatelimiter(txRatelimiter))
	accountService := accounts.NewService(cfg, accounts.NewGormStore(db), km, fc, wp, transactionService, templateService, accounts.WithTxRatelimiter(txRatelimiter))
	tokenService := tokens.NewService(cfg, tokens.NewGormStore(db), km, fc, wp, transactionService, templateService, accountService)
	opsService := ops.NewService(cfg, ops.NewGormStore(db), templateService, transactionService, tokenService)

	// Register a handler for account added events
	accounts.AccountAdded.Register(&tokens.AccountAddedHandler{
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	err = accountService.InitAdminAccount(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	wp.Start()
	log.Info("Started workerpool")

	// HTTP handling
	systemHandler := handlers.NewSystem(systemService)
	templateHandler := handlers.NewTemplates(templateService)
	jobsHandler := handlers.NewJobs(jobsService)
	accountHandler := handlers.NewAccounts(accountService)
	transactionHandler := handlers.NewTransactions(transactionService)
	tokenHandler := handlers.NewTokens(tokenService)
	opsHandler := handlers.NewOps(opsService)

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

	rv.Handle("/system/sync-account-key-count", accountHandler.SyncAccountKeyCount()).Methods(http.MethodPost)

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
		log.Info("raw transactions disabled")
	}

	// Non-custodial watchlist accounts
	rv.Handle("/watchlist/accounts", accountHandler.AddNonCustodialAccount()).Methods(http.MethodPost)                // add
	rv.Handle("/watchlist/accounts/{address}", accountHandler.DeleteNonCustodialAccount()).Methods(http.MethodDelete) // delete

	// Scripts
	rv.Handle("/scripts", transactionHandler.ExecuteScript()).Methods(http.MethodPost) // execute

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
		log.Info("fungible tokens disabled")
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
		log.Info("non-fungible tokens disabled")
	}

	// Ops
	rv.Handle("/ops/missing-fungible-token-vaults/start", opsHandler.InitMissingFungibleVaults()).Methods(http.MethodGet) // start retroactive init job
	rv.Handle("/ops/missing-fungible-token-vaults/stats", opsHandler.GetMissingFungibleVaults()).Methods(http.MethodGet)  // get number of accounts with missing fungible token vaults

	h := http.TimeoutHandler(r, cfg.ServerRequestTimeout, "request timed out")
	h = handlers.UseCors(h)
	h = handlers.UseLogging(h)
	h = handlers.UseCompress(h)

	// Setup idempotency key middleware if it's enabled
	// redis for idempotency key handling
	if !cfg.DisableIdempotencyMiddleware {
		var is handlers.IdempotencyStore
		switch cfg.IdempotencyMiddlewareDatabaseType {
		// Shared SQL/Gorm store (same as for main app)
		case handlers.IdempotencyStoreTypeShared.String():
			is = handlers.NewIdempotencyStoreGorm(db)
		// Redis, separate from app db
		case handlers.IdempotencyStoreTypeRedis.String():
			if cfg.IdempotencyMiddlewareRedisURL == "" {
				log.Fatal("idempotency middleware db set to redis but Redis URL is empty")
			}
			pool := &redis.Pool{
				MaxIdle:   80,
				MaxActive: 12000,
				Dial: func() (redis.Conn, error) {
					c, err := redis.DialURL(cfg.IdempotencyMiddlewareRedisURL)
					if err != nil {
						panic(err.Error())
					}
					return c, err
				},
			}

			client := pool.Get()

			defer func() {
				log.Info("Closing Redis client..")
				if err := client.Close(); err != nil {
					log.Warn(err)
				}
			}()

			is = handlers.NewIdempotencyStoreRedis(client)
		case handlers.IdempotencyStoreTypeLocal.String():
			is = handlers.NewIdempotencyStoreLocal()
		}

		h = handlers.UseIdempotency(h, handlers.IdempotencyHandlerOptions{
			Expiry:      1 * time.Hour,
			IgnorePaths: []string{"/v1/scripts"}, // Scripts are read-only
		}, is)
	}

	// Server boilerplate
	srv := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		WriteTimeout: 0, // Disabled, set cfg.ServerRequestTimeout instead
		ReadTimeout:  0, // Disabled, set cfg.ServerRequestTimeout instead
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.
			WithFields(log.Fields{
				"host": cfg.Host,
				"port": cfg.Port,
			}).
			Info("Server listening")
		if err := srv.ListenAndServe(); err != nil {
			log.Warn(err)
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

			token_count := len(tt)
			event_types := make([]string, token_count)

			// Listen for enabled tokens deposit events
			for i, token := range tt {
				event_types[i] = templates.DepositEventTypeFromToken(token)
			}

			return event_types, nil
		}

		listener := chain_events.NewListener(
			fc, store, getTypes,
			cfg.ChainListenerMaxBlocks,
			cfg.ChainListenerInterval,
			cfg.ChainListenerStartingHeight,
			chain_events.WithSystemService(systemService),
		)

		defer func() {
			listener.Stop()
			log.Info("Stopped chain events listener")
		}()

		// Register a handler for chain events
		chain_events.ChainEvent.Register(&tokens.ChainEventHandler{
			AccountService:  accountService,
			ChainListener:   listener,
			TemplateService: templateService,
			TokenService:    tokenService,
		})

		listener.Start()

		log.Info("Started chain events listener")
	}

	// Trap interupt or sigterm and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	log.Infof("Got signal: %s. Shutting down..", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Warnf("Error in server shutdown: %s", err)
	}
}
