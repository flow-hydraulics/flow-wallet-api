package tests

import (
	"context"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
)

func Test_Add_New_Non_Custodial_Account(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	a, err := svc.AddNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	if a.Address != addr {
		t.Fatalf("expected a.Address = %q, got %q", addr, a.Address)
	}
}

func Test_Add_Existing_Non_Custodial_Account_fails(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.AddNonCustodialAccount(context.Background(), addr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Add_Non_Custodial_Account_After_Delete(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	// One must be able to add the same account again after it was deleted.
	_, err = svc.AddNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Delete_Non_Existing_Account(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	err := svc.DeleteNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Delete_Fails_On_Custodial_Account(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	_, a, err := svc.Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(context.Background(), a.Address)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Delete_Non_Custodial_Account_Is_Idempotent(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
}
