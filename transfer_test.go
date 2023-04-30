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

// TestTransfer - create user 'from' and user 'userTo', emit amount to user 'userFrom' and transfer token from 'userFrom' to 'userTo'
func TestTransfer(t *testing.T) {
	runner.Run(t, "Emission of `fiat` token and it's transfer from user-to-user", func(t provider.T) {
		t.Severity(allure.BLOCKER)
		t.Description("Testing emitting token, and transferring it from one to another user")
		t.Tags("positive", "transfer")

		var (
			ctx = context.Background()

			issuerPrKey, userFromPrKey                    ed25519.PrivateKey
			issuerPKey, userFromPKey                      ed25519.PublicKey
			issuerPKeyStr, userFromPKeyStr, userToPKeyStr string
			userFromAddress, userToAddress                string
		)

		t.WithNewStep("Generate cryptos for users and saving it to `acl` chaincode", func(sCtx provider.StepCtx) {
			sCtx.WithNewAsyncStep("Get crypto for issuer from env and saving to `acl` chaincode", func(sCtx provider.StepCtx) {
				var err error
				issuerPrKey, issuerPKey, err = utils.GetPrivateKeyFromBase58Check(os.Getenv(utils.EnvFiatIssuerPrivateKey))
				sCtx.Assert().NoError(err)
				issuerPKeyStr = base58.Encode(issuerPKey)

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

			sCtx.WithNewAsyncStep("Generate cryptos for first user (user from) and save to `acl` chaincode", func(sCtx provider.StepCtx) {
				var err error
				userFromPrKey, userFromPKey, err = utils.GeneratePrivateAndPublicKey()
				sCtx.Assert().NoError(err)
				userFromPKeyStr = base58.Encode(userFromPKey)
				userFromAddress, err = utils.GetAddressByPublicKey(userFromPKey)
				sCtx.Assert().NoError(err)

				sCtx.WithNewStep("Add first user to `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", userFromPKeyStr, "test", "testuser", "true")
					sCtx.Assert().NoError(err)
				})

				sCtx.WithNewStep("Check first user in `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "checkKeys", userFromPKeyStr)
					sCtx.Assert().NoError(err)
				})
			})

			sCtx.WithNewAsyncStep("Generate cryptos for second user (user to) ans save to `acl` chaincode", func(sCtx provider.StepCtx) {
				_, userToPKey, err := utils.GeneratePrivateAndPublicKey()
				sCtx.Assert().NoError(err)
				userToPKeyStr = base58.Encode(userToPKey)
				userToAddress, err = utils.GetAddressByPublicKey(userToPKey)
				sCtx.Assert().NoError(err)

				sCtx.WithNewStep("Add second user to `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", userToPKeyStr, "test", "testuser", "true")
					sCtx.Assert().NoError(err)
				})

				sCtx.WithNewStep("Check second user in `acl` chaincode", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "checkKeys", userToPKeyStr)
					sCtx.Assert().NoError(err)
				})
			})
		})

		t.WithNewStep("Emit FIAT token to first user", func(sCtx provider.StepCtx) {
			var (
				emitAmount     = "1"
				signedEmitArgs []string
				err            error
			)
			sCtx.WithNewStep("Sign arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs, err = utils.Sign(issuerPrKey, issuerPKey, "fiat", "fiat", "emit", []string{userFromAddress, emitAmount})
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check balance of first user after emission", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "balanceOf", userFromAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"1\"", string(resp.Payload))
			})
		})

		t.WithNewStep("Transfer previously emitted token FIAT to second user", func(sCtx provider.StepCtx) {
			var (
				amount             = "1"
				signedTransferArgs []string
				err                error
			)

			sCtx.WithNewStep("Sign arguments before transfer process", func(sCtx provider.StepCtx) {
				signedTransferArgs, err = utils.Sign(userFromPrKey, userFromPKey, "fiat", "fiat", "transfer", []string{userToAddress, amount, "ref transfer"})
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke fiat chaincode to transfer", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "transfer", signedTransferArgs...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check balances of first and second user", func(sCtx provider.StepCtx) {
				sCtx.WithNewAsyncStep("Check balance of first user", func(sCtx provider.StepCtx) {
					resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"fiat", "balanceOf", userFromAddress)
					sCtx.Assert().NoError(err)
					sCtx.Assert().Equal("\"0\"", string(resp.Payload))
				})
				sCtx.WithNewAsyncStep("Check balance of second user", func(sCtx provider.StepCtx) {
					resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"fiat", "balanceOf", userToAddress)
					sCtx.Assert().NoError(err)
					sCtx.Assert().Equal("\"1\"", string(resp.Payload))
				})
			})
		})
	})
}
