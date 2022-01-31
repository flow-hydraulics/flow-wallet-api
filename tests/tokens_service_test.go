package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/tests/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

func Test_TokensSetup(t *testing.T) {
	cfg := test.LoadConfig(t)
	svc := test.GetServices(t, cfg).GetTokens()

	_, testAccount, err := test.GetServices(t, cfg).GetAccounts().Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	type input struct {
		sync      bool
		tokenName string
		address   string
	}
	type expect struct {
		job    *jobs.Job
		tx     *transactions.Transaction
		errMsg string
	}
	testCases := []struct {
		name   string
		input  input
		expect expect
	}{
		{
			name: "success",
			input: input{
				sync:      false,
				tokenName: "fusd",
				address:   testAccount.Address,
			},
			expect: expect{
				job: &jobs.Job{
					Type:                   transactions.TransactionJobType,
					State:                  jobs.Complete,
					Error:                  "",
					Result:                 "",
					ExecCount:              int(1),
					ShouldSendNotification: true,
				},
				tx: &transactions.Transaction{
					TransactionType: transactions.FtSetup,
					ProposerAddress: testAccount.Address,
				},
				errMsg: "",
			},
		},
		{
			name: "fail address not found",
			input: input{
				sync:      false,
				tokenName: "fusd",
				address:   "0x0ae53cb6e3f42a79",
			},
			expect: expect{
				job:    nil,
				tx:     nil,
				errMsg: "record not found",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			job, tx, err := svc.Setup(context.Background(), tc.input.sync, tc.input.tokenName, tc.input.address)
			// Wait for job to execute
			for job != nil && job.State != jobs.Complete && job.State != jobs.Failed {
				time.Sleep(10 * time.Millisecond)
			}
			jobOpts := []cmp.Option{
				cmpopts.IgnoreFields(jobs.Job{}, "TransactionID", "ExecCount"),
				cmpopts.IgnoreTypes(uuid.UUID{}, time.Time{}),
			}
			if diff := cmp.Diff(tc.expect.job, job, jobOpts...); diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
			txOpts := []cmp.Option{
				cmpopts.IgnoreFields(transactions.Transaction{}, "TransactionId", "FlowTransaction"),
				cmpopts.IgnoreTypes(time.Time{}),
			}
			if diff := cmp.Diff(tc.expect.tx, tx, txOpts...); diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if diff := cmp.Diff(tc.expect.errMsg, errMsg); diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}
}
