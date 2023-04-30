package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/tickets-dao/integration/utils"
)

func TestAclAddUser(t *testing.T) {
	runner.Run(t, "add user in acl chaincode", func(t provider.T) {
		ctx := context.Background()
		t.Severity(allure.BLOCKER)
		t.Description("As member of organization add user in acl chaincode and validate that user was added")
		t.Tags("smoke", "acl", "positive")

		var publicKey string

		t.WithNewStep("Generate cryptos for user", func(sCtx provider.StepCtx) {
			_, pkey, err := utils.GeneratePrivateAndPublicKey()
			sCtx.Assert().NoError(err)
			publicKey = base58.Encode(pkey)
		})

		t.WithNewStep("Add user by invoking method `addUser` of chaincode `acl` with valid parameters", func(sCtx provider.StepCtx) {
			_, err := utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
				os.Getenv(utils.EnvHlfProxyAuthToken),
				"acl", "addUser", publicKey, "test", "testuser", "true")
			sCtx.Assert().NoError(err)
		})

		time.Sleep(utils.BatchTransactionTimeout)
		t.WithNewStep("Check user is created by querying method `checkKeys` of chaincode `acl`", func(sCtx provider.StepCtx) {
			_, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
				os.Getenv(utils.EnvHlfProxyAuthToken), "acl", "checkKeys", publicKey)
			sCtx.Assert().NoError(err)
		})
	})
}
