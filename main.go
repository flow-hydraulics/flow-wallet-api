package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

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

func main() {
	godotenv.Load(".env")

	// Configs
	// TODO: put these in a separate package
	var (
		wait                 time.Duration
		disable_account_mgmt bool
		disable_raw_tx       bool
		disable_ft           bool
		// disable_nft   bool
		flow_gateway   string = "localhost:3569" // TODO
		encryption_key string = "1234"           // TODO
	)
	s_acc_addr := flow.HexToAddress(os.Getenv("SERVICE_ACC_ADDRESS"))
	s_key_type := os.Getenv("SERVICE_ACC_KEY_TYPE")
	s_key_value := os.Getenv("SERVICE_ACC_KEY_VALUE")
	s_key_idx, err := strconv.ParseUint(os.Getenv("SERVICE_ACC_KEY_INDEX"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.BoolVar(&disable_account_mgmt, "disable-account-mgmt", false, "disable account management")
	flag.BoolVar(&disable_raw_tx, "disable-raw-tx", false, "disable sending raw transactions for an account")
	flag.BoolVar(&disable_ft, "disable-ft", false, "disable fungible token functionality")
	flag.Parse()

	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	fc, err := client.New(flow_gateway, grpc.WithInsecure())
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
			AccountAddress: s_acc_addr,
			Index:          int(s_key_idx),
			Type:           s_key_type,
			Value:          s_key_value,
		},
		encryption_key,
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
		Addr:         ":3000",
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
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
