package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/tests/test"
)

func Test_Add_New_Non_Custodial_Account(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	a, err := svc.AddNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	if a.Address != addr {
		t.Fatalf("expected a.Address = %q, got %q", addr, a.Address)
	}
}

func Test_Add_Existing_Non_Custodial_Account_fails(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.AddNonCustodialAccount(addr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Add_Non_Custodial_Account_After_Delete(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	// One must be able to add the same account again after it was deleted.
	_, err = svc.AddNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Delete_Non_Existing_Account(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	err := svc.DeleteNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Delete_Fails_On_Custodial_Account(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	_, a, err := svc.Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(a.Address)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Delete_Non_Custodial_Account_Is_Idempotent(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetAccounts()

	addr := "0x0123456789"

	_, err := svc.AddNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.DeleteNonCustodialAccount(addr)
	if err != nil {
		t.Fatal(err)
	}
}

// Test if the service is able to concurrently create multiple accounts
func Test_Add_Multiple_New_Custodial_Accounts(t *testing.T) {
	t.Skip("sqlite will cause a database locked error")

	cfg := test.LoadConfig(t)

	instanceCount := 1
	accountsToCreate := instanceCount * 5
	// Worst case scenario where theoretically maximum number of transactions are done concurrently
	cfg.WorkerCount = uint(accountsToCreate / instanceCount)

	svcs := make([]test.Services, instanceCount)

	for i := 0; i < instanceCount; i++ {
		svcs[i] = test.GetServices(t, cfg)
	}

	if cfg.AdminProposalKeyCount <= 1 {
		t.Skip("skipped as \"cfg.AdminProposalKeyCount\" is less than or equal to 1")
	}

	if accounts, err := svcs[0].GetAccounts().List(0, 0); err != nil {
		t.Fatal(err)
	} else if len(accounts) > 1 {
		t.Fatal("expected there to be only 1 account")
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error, accountsToCreate*2)

	for i := 0; i < accountsToCreate; i++ {
		wg.Add(1)
		go func(i int, svcs []test.Services) {
			defer wg.Done()

			svc := svcs[i%instanceCount].GetAccounts()
			jobSvc := svcs[i%instanceCount].GetJobs()

			job, _, err := svc.Create(context.Background(), false)
			if err != nil {
				errChan <- err
				return
			}

			if _, err := test.WaitForJob(jobSvc, job.ID.String()); err != nil {
				errChan <- err
			}
		}(i, svcs)
	}

	wg.Wait()

	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}

	if accounts, err := svcs[0].GetAccounts().List(0, 0); err != nil {
		t.Fatal(err)
	} else if len(accounts) < 1+accountsToCreate {
		t.Fatalf("expected there to be %d accounts", 1+accountsToCreate)
	}
}
