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
	"golang.org/x/crypto/ed25519"
)

// TestTxTTLSuccess - create user 'from', emit amount to user 'userFrom'
func TestTxTTLSuccess(t *testing.T) {
	runner.Run(t, "Emission of `fiat` token and check txTtl", func(t provider.T) {
		t.Severity(allure.BLOCKER)
		t.Description("Testing check txTtl")
		t.Tags("positive", "tx ttl")

		var (
			ctx = context.Background()

			issuerPrivateKey         ed25519.PrivateKey
			issuerPubKey, userPubKey ed25519.PublicKey
		)

		t.WithNewStep("Generate cryptos for users and saving it to `acl` chaincode", func(sCtx provider.StepCtx) {
			sCtx.WithNewAsyncStep("Get crypto for issuer from env and saving to `acl` chaincode", func(sCtx provider.StepCtx) {
				var err error
				issuerPrivateKey, issuerPubKey, err = utils.GetPrivateKeyFromBase58Check(os.Getenv(utils.EnvFiatIssuerPrivateKey))
				sCtx.Assert().NoError(err)
				issuerPKeyStr := base58.Encode(issuerPubKey)

				sCtx.WithNewStep("Add issuer user to `acl` chaincode", func(sCtx provider.StepCtx) {
					_, _ = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", issuerPKeyStr, "test", "testuser", "true")
				})
				sCtx.WithNewStep("Check issuer user in `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "checkKeys", issuerPKeyStr)
					sCtx.Assert().NoError(err)
				})
			})

			sCtx.WithNewAsyncStep("Generate cryptos for user and save to `acl` chaincode", func(sCtx provider.StepCtx) {
				var err error
				_, userPubKey, err = utils.GeneratePrivateAndPublicKey()
				sCtx.Assert().NoError(err)
				userPubKeyStr := base58.Encode(userPubKey)

				sCtx.WithNewStep("Add user to `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", userPubKeyStr, "test", "testuser", "true")
					sCtx.Assert().NoError(err)
				})

				sCtx.WithNewStep("Check user in `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "checkKeys", userPubKeyStr)
					sCtx.Assert().NoError(err)
				})
			})
		})

		t.WithNewStep("Emit FIAT token to user", func(sCtx provider.StepCtx) {
			var (
				emitAmount     = "1"
				signedEmitArgs []string
				err            error
				resp           *utils.Response
			)

			userAddress, err := utils.GetAddressByPublicKey(userPubKey)
			sCtx.Assert().NoError(err)

			sCtx.WithNewStep("Sign arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					"fiat",
					"fiat",
					"emit",
					[]string{userAddress, emitAmount},
				)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check balance of user after emission", func(sCtx provider.StepCtx) {
				resp, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "balanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().NotNil(resp)
				sCtx.Assert().Equal("\"1\"", string(resp.Payload))
			})
		})
	})
}
