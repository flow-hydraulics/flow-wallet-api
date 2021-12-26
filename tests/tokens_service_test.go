package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

func Test_TokensSetup(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svc := test.GetServices(t, cfg).GetTokens()

	type input struct {
		sync      bool
		tokenName string
		address   string
	}
	type expect struct {
		job *jobs.Job
		tx  *transactions.Transaction
		err bool
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
				tokenName: "flowToken",
				address:   cfg.AdminAddress,
			},
			expect: expect{
				job: &jobs.Job{
					Type:                   transactions.TransactionJobType,
					State:                  jobs.Init,
					Error:                  "",
					Result:                 "",
					ExecCount:              int(1),
					ShouldSendNotification: false,
				},
				tx: &transactions.Transaction{
					TransactionType: transactions.FtSetup,
					ProposerAddress: cfg.AdminAddress,
				},
				err: false,
			},
		},
		{
			name: "fail address not found",
			input: input{
				sync:      false,
				tokenName: "flowToken",
				address:   "0x0ae53cb6e3f42a79",
			},
			expect: expect{
				job: nil,
				tx:  nil,
				err: true,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			job, tx, err := svc.Setup(context.Background(), tc.input.sync, tc.input.tokenName, tc.input.address)
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
			if !tc.expect.err && err != nil {
				t.Fatal(err)
			}
		})
	}
}
