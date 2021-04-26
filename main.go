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

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	fc, err := client.New("localhost:3569", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	db, err := memory.NewDataStore()
	if err != nil {
		log.Fatal(err)
	}

	s_key_idx, err := strconv.ParseUint(os.Getenv("SERVICE_INDEX"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	ks, err := simple.NewKeyStore(
		db,
		store.AccountKey{
			AccountAddress: flow.HexToAddress(os.Getenv("SERVICE_ADDRESS")),
			Index:          int(s_key_idx),
			Type:           os.Getenv("SERVICE_TYPE"),
			Value:          os.Getenv("SERVICE_VALUE"),
		},
		"1234",
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
	ra.HandleFunc("", accounts.List)
	ra.HandleFunc("", accounts.Create).Methods("POST")

	// Handle "/accounts/{address}"
	raa := ra.PathPrefix("/{address}").Subrouter()
	raa.HandleFunc("", accounts.Details).Methods("GET")
	raa.HandleFunc("", accounts.Update).Methods("PUT")
	raa.HandleFunc("", accounts.Delete).Methods("DELETE")
	raa.HandleFunc("/transactions", transactions.SendTransaction).Methods("POST")

	// Handle "/accounts/{address}/fungible-tokens/{tokenName}"
	rft := raa.PathPrefix("/fungible-tokens/{tokenName}").Subrouter()
	rft.HandleFunc("", fungibleTokens.Details).Methods("GET")
	rft.HandleFunc("", fungibleTokens.Init).Methods("POST")
	rft.HandleFunc("/withdrawals", fungibleTokens.ListWithdrawals).Methods("GET")
	rft.HandleFunc("/withdrawals", fungibleTokens.CreateWithdrawal).Methods("POST")
	rft.HandleFunc("/withdrawals/{transactionId}", fungibleTokens.WithdrawalDetails).Methods("GET")

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
