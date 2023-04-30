package integration

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/tickets-dao/integration/utils"
	"golang.org/x/crypto/ed25519"
)

const itSymbol = "industrial"

// TestNonceTTLNotZero - create user 'from', multiple emit amount to user 'userFrom'
func TestNonceTTLNotZero(t *testing.T) {
	runner.Run(t, "Emission of `fiat` token and check nonce ttl", func(t provider.T) {
		t.Severity(allure.BLOCKER)
		t.Description("Testing check nonce ttl")
		t.Tags("positive", "nonce ttl")

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
					_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", issuerPKeyStr, "test", "testuser", "true")
					sCtx.Assert().True(err == nil || strings.Contains(err.Error(), "already exists"))
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
				emitAmount      = "1"
				signedEmitArgs1 []string
				signedEmitArgs2 []string
				signedEmitArgs3 []string
				signedEmitArgs4 []string
				err             error
				resp            *utils.Response
			)

			userAddress, err := utils.GetAddressByPublicKey(userPubKey)
			sCtx.Assert().NoError(err)

			sCtx.WithNewStep("Sign 1 arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs1, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					"fiat",
					"fiat",
					"emit",
					[]string{userAddress, emitAmount},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.MoreNonceTTL)

			sCtx.WithNewStep("Sign 2 arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs2, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					"fiat",
					"fiat",
					"emit",
					[]string{userAddress, emitAmount},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(time.Millisecond)

			sCtx.WithNewStep("Sign 3 arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs3, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					"fiat",
					"fiat",
					"emit",
					[]string{userAddress, emitAmount},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(time.Millisecond)

			sCtx.WithNewStep("Sign 4 arguments before emission process", func(sCtx provider.StepCtx) {
				signedEmitArgs4, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					"fiat",
					"fiat",
					"emit",
					[]string{userAddress, emitAmount},
				)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 3 fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs3...)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 2 fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs2...)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 1 fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs1...)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke again 3 fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs3...)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 4 fiat chaincode by issuer for token emission", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "emit", signedEmitArgs4...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check balance of user after emission", func(sCtx provider.StepCtx) {
				resp, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					"fiat", "balanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().NotNil(resp)
				sCtx.Assert().Equal("\"3\"", string(resp.Payload))
			})
		})
	})
}

// TestNonceTTLZero - create user 'from', multiple emit amount to user 'userFrom'
func TestNonceTTLZero(t *testing.T) {
	runner.Run(t, "Emission of `industrial` token and check nonce ttl", func(t provider.T) {
		t.Severity(allure.BLOCKER)
		t.Description("Testing check nonce ttl")
		t.Tags("positive", "nonce ttl")

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
					_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
						os.Getenv(utils.EnvHlfProxyAuthToken),
						"acl", "addUser", issuerPKeyStr, "test", "testuser", "true")
					sCtx.Assert().True(err == nil || strings.Contains(err.Error(), "already exists"))
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

		t.WithNewStep("Emit industrial token to user", func(sCtx provider.StepCtx) {
			var (
				emitAmount         = "1"
				groupId            = "202010"
				signedEmitArgsInit []string
				signedEmitArgs1    []string
				signedEmitArgs2    []string
				signedEmitArgs3    []string
				err                error
				resp               *utils.Response
			)

			userAddress, err := utils.GetAddressByPublicKey(userPubKey)
			sCtx.Assert().NoError(err)

			sCtx.WithNewStep("Sign initialize tx", func(sCtx provider.StepCtx) {
				signedEmitArgsInit, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					itSymbol,
					itSymbol,
					"initialize",
					[]string{},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Invoke Init it chaincode", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "initialize", signedEmitArgsInit...)
				sCtx.Assert().NoError(err)
			})

			// issuer.SignedInvoke("it", "transferIndustrial", user.Address(), "202101", "100", "")
			sCtx.WithNewStep("Sign 1 arguments before transfer process", func(sCtx provider.StepCtx) {
				signedEmitArgs1, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					itSymbol,
					itSymbol,
					"transferIndustrial",
					[]string{userAddress, groupId, emitAmount, ""},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(time.Millisecond)

			sCtx.WithNewStep("Sign 2 arguments before transfer process", func(sCtx provider.StepCtx) {
				signedEmitArgs2, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					itSymbol,
					itSymbol,
					"transferIndustrial",
					[]string{userAddress, groupId, emitAmount, ""},
				)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(time.Millisecond)

			sCtx.WithNewStep("Sign 3 arguments before transfer process", func(sCtx provider.StepCtx) {
				signedEmitArgs3, err = utils.Sign(
					issuerPrivateKey,
					issuerPubKey,
					itSymbol,
					itSymbol,
					"transferIndustrial",
					[]string{userAddress, groupId, emitAmount, ""},
				)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 2 it chaincode by issuer for token transfer", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "transferIndustrial", signedEmitArgs2...)
				sCtx.Assert().NoError(err)
			})

			sCtx.WithNewStep("Invoke 1 it chaincode by issuer for token transfer", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "transferIndustrial", signedEmitArgs1...)
				sCtx.Assert().Contains(err.Error(), "incorrect nonce")
			})

			sCtx.WithNewStep("Invoke again 2 it chaincode by issuer for token transfer", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "transferIndustrial", signedEmitArgs2...)
				sCtx.Assert().Contains(err.Error(), "incorrect nonce")
			})

			sCtx.WithNewStep("Invoke 3 it chaincode by issuer for token transfer", func(sCtx provider.StepCtx) {
				_, err = utils.Invoke(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "transferIndustrial", signedEmitArgs3...)
				sCtx.Assert().NoError(err)
			})

			time.Sleep(utils.BatchTransactionTimeout)
			sCtx.WithNewStep("Check balance of user after transferIndustrial", func(sCtx provider.StepCtx) {
				resp, err = utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken),
					itSymbol, "industrialBalanceOf", userAddress)
				sCtx.Assert().NoError(err)
				sCtx.Assert().NotNil(resp)
				var balances map[string]string
				sCtx.Assert().NoError(json.Unmarshal(resp.Payload, &balances))

				balance, ok := balances[groupId]
				sCtx.Assert().True(ok)
				sCtx.Assert().Equal("2", balance)
			})
		})
	})
}
