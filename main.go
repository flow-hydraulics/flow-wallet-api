package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/handlers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store/postgres"
	"github.com/gorilla/mux"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	db, err := postgres.NewDataStore()
	if err != nil {
		log.Fatal(err)
	}

	ks, err := postgres.NewKeyStore()
	if err != nil {
		log.Fatal(err)
	}

	// sAcct, err := account.NewFromFlowFile("./flow.json", "emulator-account")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fc, err := client.New("localhost:3569", grpc.WithInsecure())
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
