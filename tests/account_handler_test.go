package tests

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/handlers"
	"github.com/flow-hydraulics/flow-wallet-api/tests/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/gorilla/mux"
	"github.com/onflow/flow-go-sdk"
)

func TestEmulatorAcceptsSignedTransaction(t *testing.T) {
	cfg := test.LoadConfig(t)
	svcs := test.GetServices(t, cfg)

	accHandler := handlers.NewAccounts(svcs.GetAccounts())
	txHandler := handlers.NewTransactions(svcs.GetTransactions())

	router := mux.NewRouter()
	router.Handle("/", accHandler.Create()).Methods(http.MethodPost)
	router.Handle("/{address}/sign", txHandler.Sign()).Methods(http.MethodPost)

	// Create signing account.
	var account accounts.Account
	res := send(router, http.MethodPost, "/?sync=true", nil)
	assertStatusCode(t, res, http.StatusCreated)
	fromJsonBody(t, res, &account)

	// Transaction:
	code := "transaction(greeting: String) { prepare(signer: AuthAccount){} execute { log(greeting.concat(\", World!\")) }}"
	args := "[{\"type\":\"String\",\"value\":\"Hello\"}]"

	// Sign it.
	body := bytes.NewBufferString(fmt.Sprintf("{\"code\":%q,\"arguments\":%s}", code, args))
	res = send(router, http.MethodPost, fmt.Sprintf("/%s/sign", account.Address), body)
	assertStatusCode(t, res, http.StatusCreated)

	var txResp transactions.SignedTransactionJSONResponse
	fromJsonBody(t, res, &txResp)

	tx := flow.NewTransaction().
		SetScript([]byte(txResp.Code)).
		SetReferenceBlockID(flow.HexToID(txResp.ReferenceBlockID)).
		SetGasLimit(txResp.GasLimit).
		SetProposalKey(flow.HexToAddress(txResp.ProposalKey.Address), txResp.ProposalKey.KeyIndex, txResp.ProposalKey.SequenceNumber).
		SetPayer(flow.HexToAddress(txResp.Payer))

	for _, arg := range txResp.Arguments {
		tx.AddRawArgument(arg) // nolint
	}

	for _, a := range txResp.Authorizers {
		tx.AddAuthorizer(flow.HexToAddress(a))
	}

	for _, s := range txResp.PayloadSignatures {
		bs, err := hex.DecodeString(s.Signature)
		if err != nil {
			t.Fatal(err)
		}
		tx.AddPayloadSignature(flow.HexToAddress(s.Address), s.KeyIndex, bs)
	}

	for _, s := range txResp.EnvelopeSignatures {
		bs, err := hex.DecodeString(s.Signature)
		if err != nil {
			t.Fatal(err)
		}
		tx.AddEnvelopeSignature(flow.HexToAddress(s.Address), s.KeyIndex, bs)
	}

	ctx := context.Background()
	client := test.NewFlowClient(t, cfg)
	_, err := flow_helpers.SendAndWait(ctx, client, *tx, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatchlistAccountManagement(t *testing.T) {
	cfg := test.LoadConfig(t)
	fc := test.NewFlowClient(t, cfg)
	svcs := test.GetServices(t, cfg)
	km := svcs.GetKeyManager()

	accHandler := handlers.NewAccounts(svcs.GetAccounts())

	router := mux.NewRouter()
	router.Handle("/", accHandler.AddNonCustodialAccount()).Methods(http.MethodPost)
	router.Handle("/{address}", accHandler.Details()).Methods(http.MethodGet)
	router.Handle("/{address}", accHandler.DeleteNonCustodialAccount()).Methods(http.MethodDelete)

	// Create a non-custodial account.
	adminAuthorizer, err := km.AdminAuthorizer(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	nonCustodialAccount := test.NewFlowAccount(t, fc, adminAuthorizer.Address, adminAuthorizer.Key, adminAuthorizer.Signer)

	// Add created non-custodial account to watchlist.
	account := accounts.Account{Address: nonCustodialAccount.Address.Hex()}
	buf := bytes.NewBuffer(asJson(&account))
	res := send(router, http.MethodPost, "/", buf)
	assertStatusCode(t, res, http.StatusCreated)
	fromJsonBody(t, res, &account)

	// Ensure that account can be found.
	res = send(router, http.MethodGet, fmt.Sprintf("/%s", account.Address), nil)
	assertStatusCode(t, res, http.StatusOK)
	fromJsonBody(t, res, &account)

	if account.Address != flow_helpers.FormatAddress(nonCustodialAccount.Address) {
		t.Fatalf("read account address doesn't match - expected %q, got %q", flow_helpers.FormatAddress(nonCustodialAccount.Address), account.Address)
	}

	if account.Type != accounts.AccountTypeNonCustodial {
		t.Fatalf("read account type doesn't match - expected %q, got %q", accounts.AccountTypeNonCustodial, account.Type)
	}

	// Remove the non-custodial account from watchlist.
	res = send(router, http.MethodDelete, fmt.Sprintf("/%s", account.Address), nil)
	assertStatusCode(t, res, http.StatusOK)

	// Ensure that it's not found anymore.
	res = send(router, http.MethodGet, fmt.Sprintf("/%s", account.Address), nil)
	assertStatusCode(t, res, http.StatusNotFound)
}

func assertStatusCode(t *testing.T, res *http.Response, expected int) {
	t.Helper()
	if res.StatusCode != expected {
		bs, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		t.Fatalf("expected HTTP response status code %d, got %d: %s", expected, res.StatusCode, string(bs))
	}
}

func asJson(v interface{}) []byte {
	if v == nil {
		return nil
	}
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bs
}

func fromJsonBody(t *testing.T, res *http.Response, v interface{}) {
	t.Helper()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(bs, v)
	if err != nil {
		t.Fatal(err)
	}
}

func send(router *mux.Router, method, path string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}
