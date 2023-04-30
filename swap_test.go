package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
	"github.com/tickets-dao/integration/utils"
	"golang.org/x/crypto/ed25519"
)

const (
	DefaultSwapHash = "7d4e3eec80026719639ed4dba68916eb94c7a49a053e05c8f9578fe4e5a3d7ea"
	DefaultSwapKey  = "12345"
)

// TestSwap - create user, emit amount to fiat, swap amount from fiat channel to cc channel, check amount is moved
func TestSwap(t *testing.T) {
	// skipped due to refactoring
	runner.Run(t, "swap token from fiat to cc and swap back", func(t provider.T) {
		ctx := context.Background()
		t.Severity(allure.BLOCKER)
		t.Description("Acceptance of emitting amount to fiat and swap amount from fiat channel to cc channel")
		t.Tags("positive", "swap")

		issuerPrivateKey, issuerPublicKey, err := utils.GetPrivateKeyFromBase58Check(os.Getenv(utils.EnvFiatIssuerPrivateKey))
		assert.NoError(t.RealT(), err)

		issuerEd25519PublicKeyBase58 := base58.Encode(issuerPublicKey)
		t.WithNewStep("addUser. Try to add issuer user in acl, no handle err because issuer can exist", func(sCtx provider.StepCtx) {
			_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
				os.Getenv(utils.EnvHlfProxyAuthToken), "acl", "addUser", issuerEd25519PublicKeyBase58, "test", "testuser", "true")
			sCtx.Assert().True(err == nil || strings.Contains(err.Error(), "already exists"))
		})

		t.WithNewStep("Check issuer public key after adding", func(sCtx provider.StepCtx) {
			_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
				os.Getenv(utils.EnvHlfProxyAuthToken), "acl", "checkKeys", issuerEd25519PublicKeyBase58)
			sCtx.Assert().NoError(err)
		})

		var userPublicKeyStr, userAddress string
		var userPrivateKey ed25519.PrivateKey
		var userPublicKey ed25519.PublicKey

		t.WithNewStep("Adding user with method `addUser` of chaincode `acl`", func(sCtx provider.StepCtx) {
			sCtx.WithNewStep("Creating cryptos for new user", func(sCtx provider.StepCtx) {
				userPrivateKey, userPublicKey, err = utils.GeneratePrivateAndPublicKey()
				t.Assert().NoError(err)
				userPublicKeyStr = base58.Encode(userPublicKey)
				userAddress, err = utils.GetAddressByPublicKey(userPublicKey)
			})

			sCtx.WithNewStep("Add user to chaincode `acl` by invoking method `addUser`", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "acl", "addUser", userPublicKeyStr, "test", "testuser", "true")
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Check user public key", func(sCtx provider.StepCtx) {
				_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "acl", "checkKeys", userPublicKeyStr)
				sCtx.Assert().NoError(err)
			})
		})

		t.WithNewStep("Emission of FIAT token to recently added user", func(sCtx provider.StepCtx) {
			var (
				amount     = "1"
				signedArgs []string
			)
			sCtx.WithNewStep("Sign arguments before sending to chaincode", func(sCtx provider.StepCtx) {
				signedArgs, err = utils.Sign(issuerPrivateKey, issuerPublicKey, "fiat", "fiat", "emit", []string{userAddress, amount})
				t.Assert().NoError(err)
			})

			sCtx.WithNewStep("Emit token", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "emit", signedArgs...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check FIAT token balance in `fiat` channel", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "balanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"1\"", string(resp.Payload))
			})
		})

		t.WithNewStep("Swap token FIAT from `fiat` channel to `cc` channel", func(sCtx provider.StepCtx) {
			swapFromTokenName := FiatName
			swapToChannel := "CC"
			swapAmount := "1"

			swapBeginArgs := []string{swapFromTokenName, swapToChannel, swapAmount, DefaultSwapHash}
			var signedSwapBeginArgs []string
			sCtx.WithNewStep("Sign arguments before swap process", func(sCtx provider.StepCtx) {
				signedSwapBeginArgs, err = utils.Sign(userPrivateKey, userPublicKey, "fiat", "fiat", "swapBegin", swapBeginArgs)
				sCtx.Assert().NoError(err)
			})

			var swapBeginTxID string
			sCtx.WithNewStep("Invoke `swapBegin` of `fiat` chaincode", func(sCtx provider.StepCtx) {
				resp, err := utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "swapBegin", signedSwapBeginArgs...)
				sCtx.Assert().NoError(err)
				swapBeginTxID = resp.TransactionID
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Invoke `swapGet` of `fiat` chaincode", func(sCtx provider.StepCtx) {
				_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "swapGet", swapBeginTxID)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Get balance of user by invoking method `balanceOf` of chaincode `cc`", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "balanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"0\"", string(resp.Payload))
			})

			sCtx.WithNewStep("Get allowed balance in `cc` channel", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "allowedBalanceOf", userAddress, "FIAT")
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"0\"", string(resp.Payload))
			})

			sCtx.WithNewStep("Stop swapping process", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "swapDone", swapBeginTxID, DefaultSwapKey)
				sCtx.Assert().NoError(err)
			})
		})

		time.Sleep(utils.BatchTransactionTimeout)
		t.WithNewStep("Check balances in channels", func(sCtx provider.StepCtx) {
			sCtx.WithNewAsyncStep("Check balance in `fiat` channel", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "balanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"0\"", string(resp.Payload))
			})
			sCtx.WithNewAsyncStep("Check balance in `cc` channel", func(sCtx provider.StepCtx) {
				resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "allowedBalanceOf", userAddress, "FIAT")
				sCtx.Assert().NoError(err)
				sCtx.Assert().Equal("\"1\"", string(resp.Payload))
			})
		})

		t.WithNewStep("Back swap from cc to fiat", func(sCtx provider.StepCtx) {
			backSwapFromTokenName := FiatName
			backSwapToChannel := FiatName
			backSwapAmount := "1"

			backSwapBeginArgs := []string{backSwapFromTokenName, backSwapToChannel, backSwapAmount, DefaultSwapHash}
			var signedBackSwapBeginArgs []string
			sCtx.WithNewStep("Sign arguments before swap process", func(sCtx provider.StepCtx) {
				signedBackSwapBeginArgs, err = utils.Sign(userPrivateKey, userPublicKey, "cc", "cc", "swapBegin", backSwapBeginArgs)
				sCtx.Assert().NoError(err)
			})

			var swapBackBeginTxID string
			sCtx.WithNewStep("Invoke `swapBegin` method of chaincode `cc`", func(sCtx provider.StepCtx) {
				resp, err := utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "swapBegin", signedBackSwapBeginArgs...)
				sCtx.Assert().NoError(err)
				swapBackBeginTxID = resp.TransactionID
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Query swaps of chaincodes in channels `fiat` and `cc`", func(sCtx provider.StepCtx) {
				sCtx.WithNewAsyncStep("Query method `swapGet` of chaincode `fiat`", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "swapGet", swapBackBeginTxID)
					sCtx.Assert().NoError(err)
				})

				sCtx.WithNewAsyncStep("Query method of chaincode `swapGet` of chaincode `cc` ", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "swapGet", swapBackBeginTxID)
					sCtx.Assert().NoError(err)
				})
			})

			sCtx.WithNewStep("Get balances in certain channels in channels `fiat` and `cc`", func(sCtx provider.StepCtx) {
				sCtx.WithNewAsyncStep("Query method `swapGet` of chaincode `fiat`", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "swapGet", swapBackBeginTxID)
					sCtx.Assert().NoError(err)
				})

				sCtx.WithNewAsyncStep("Query method of chaincode `swapGet` of chaincode `cc` ", func(sCtx provider.StepCtx) {
					_, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "swapGet", swapBackBeginTxID)
					sCtx.Assert().NoError(err)
				})
			})

			sCtx.WithNewStep("Finish swap process with method `swapDone` of chaincode `fiat`", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "swapDone", swapBackBeginTxID, DefaultSwapKey)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Get allowed balances if channels `fiat` and `cc`", func(sCtx provider.StepCtx) {
				sCtx.WithNewAsyncStep("Get allowed balance in `fiat` channel", func(sCtx provider.StepCtx) {
					resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "fiat", "balanceOf", userAddress)
					sCtx.Assert().NoError(err)
					sCtx.Assert().NotNil(resp)
					sCtx.Assert().Equal("\"1\"", string(resp.Payload))
				})
				sCtx.WithNewAsyncStep("Get allowed balance in `cc` channel", func(sCtx provider.StepCtx) {
					resp, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken), "cc", "allowedBalanceOf", userAddress, "fiat")
					sCtx.Assert().NoError(err)
					sCtx.Assert().NotNil(resp)
					sCtx.Assert().Equal("\"0\"", string(resp.Payload))
				})
			})
		})
	})
}
