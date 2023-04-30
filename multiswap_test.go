package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
	"github.com/tickets-dao/integration/utils"
)

const FiatName = "FIAT"

// TestMultiSwap - create user, emit amount to fiat, multi swap amount from fiat channel to cc channel, check amount is moved
func TestMultiSwap(t *testing.T) { // relates issue github.com/tickets-dao/foundation/-/issues/48"
	// skipped due to refactoring
	t.Skip()
	runner.Run(t, "multiswap token from fiat to cc and multiswap back", func(t provider.T) {
		t.Skip("reason: https://github.com/tickets-dao/foundation/-/issues/48")
		t.Tags("positive", "multiswap")

		t.NewStep("Prepare service to connect with API hlf proxy service")
		hlfProxy := utils.NewHlfProxyService(os.Getenv(utils.EnvHlfProxyURL), os.Getenv(utils.EnvHlfProxyAuthToken))

		t.NewStep("addUser. Get 'private key' and 'public key' Issuer user from 'ed25519 private key in format base 58 check'")
		issuerFiatEd25519PrivateKey, issuerFiatEd25519PublicKey, err := utils.GetPrivateKeyFromBase58Check(os.Getenv(utils.EnvFiatIssuerPrivateKey))
		assert.NoError(t, err)
		issuerEd25519PublicKeyBase58 := base58.Encode(issuerFiatEd25519PublicKey)
		t.NewStep("addUser. Try to add issuer user in acl, no handle err because issuer may already exist")
		_, err = hlfProxy.Invoke("acl", "addUser", issuerEd25519PublicKeyBase58, "test", "testuser", "true")
		assert.NoError(t, err)

		t.NewStep("after add user. check issuer public key")
		_, err = hlfProxy.Query("acl", "checkKeys", issuerEd25519PublicKeyBase58)
		t.Assert().NoError(err)

		t.NewStep("addUser. Generate private key for user")
		userEd25519PrivateKey, userEd25519PublicKey, err := utils.GeneratePrivateAndPublicKey()
		assert.NoError(t, err)

		userPublicKeyBase58 := base58.Encode(userEd25519PublicKey)
		userAddressBase58Check, err := utils.GetAddressByPublicKey(userEd25519PublicKey)
		assert.NoError(t, err)

		t.NewStep("addUser. invoke chaincode acl with method addUser, and create user")
		_, err = hlfProxy.Invoke("acl", "addUser", userPublicKeyBase58, "test", "testuser", "true")
		assert.NoError(t, err)

		t.NewStep("after add user. check user public key")
		_, err = hlfProxy.Query("acl", "checkKeys", userPublicKeyBase58)
		t.Assert().NoError(err)

		t.NewStep("Emit 1 FIAT token to user")
		emitAmount := "1"
		emitArgs := []string{userAddressBase58Check, emitAmount}
		signedEmitArgs, err := utils.Sign(issuerFiatEd25519PrivateKey, issuerFiatEd25519PublicKey, "fiat", "fiat", "emit", emitArgs)
		assert.NoError(t, err)
		_, err = hlfProxy.Invoke("fiat", "emit", signedEmitArgs...)
		assert.NoError(t, err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("After emit need to check balance FIAT token in fiat channel by user address")
		resp, err := hlfProxy.Query("fiat", "balanceOf", userAddressBase58Check)
		assert.NoError(t, err)
		assert.Equal(t, "\"1\"", string(resp.Payload))

		t.NewStep("Start multi swap process with call method multiSwapBegin in fiat channel. We start to move 1 FIAT token from 'fiat' channel to 'cc' channel")
		tokenPrefix := FiatName
		multiSwapAmount := "1"
		assets := fmt.Sprintf("{\"Assets\":[{\"group\":\"%s\",\"amount\":\"%s\"}]}", tokenPrefix, multiSwapAmount)
		channelTo := "CC"
		multiSwapBeginArgs := []string{tokenPrefix, assets, channelTo, DefaultSwapHash}
		signedMultiSwapBeginArgs, err := utils.Sign(userEd25519PrivateKey, userEd25519PublicKey, "fiat", "fiat", "multiSwapBegin", multiSwapBeginArgs)
		assert.NoError(t, err)
		multiSwapBeginResp, err := hlfProxy.Invoke("fiat", "multiSwapBegin", signedMultiSwapBeginArgs...)
		assert.NoError(t, err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("By transaction id from response multiSwapBegin need to check multi swap record in fiat channel")
		_, err = hlfProxy.Query("fiat", "multiSwapGet", multiSwapBeginResp.TransactionID)
		assert.NoError(t, err)

		t.NewStep("By transaction id from response multiSwapBegin need to check multi swap record in cc channel")
		_, err = hlfProxy.Query("cc", "multiSwapGet", multiSwapBeginResp.TransactionID)
		assert.NoError(t, err)

		t.NewStep("After multiSwapBegin need to check balance FIAT token in fiat channel by user address. This balance must change")
		resp, err = hlfProxy.Query("fiat", "balanceOf", userAddressBase58Check)
		assert.NoError(t, err)
		assert.Equal(t, "\"0\"", string(resp.Payload))

		t.NewStep("After multiSwapBegin need to check allowed balance FIAT token in fiat channel by user address.")
		resp, err = hlfProxy.Query("cc", "allowedBalanceOf", userAddressBase58Check, "FIAT")
		assert.NoError(t, err)
		assert.Equal(t, "\"0\"", string(resp.Payload))

		t.NewStep("Complete multi swap process. Invoke multiSwapDone")
		_, err = hlfProxy.Invoke("cc", "multiSwapDone", multiSwapBeginResp.TransactionID, DefaultSwapKey)
		assert.NoError(t, err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("After multiSwapDone need to check balance FIAT token in fiat channel by user address.")
		resp, err = hlfProxy.Query("fiat", "balanceOf", userAddressBase58Check)
		assert.NoError(t, err)
		assert.Equal(t, "\"0\"", string(resp.Payload))

		t.NewStep("After multiSwapDone need to check allowed balance FIAT token in fiat channel by user address. This balance must change")
		resp, err = hlfProxy.Query("cc", "allowedBalanceOf", userAddressBase58Check, "FIAT")
		assert.NoError(t, err)
		assert.Equal(t, "\"1\"", string(resp.Payload))

		t.NewStep("Begin multiswap - back FIAT token from cc to fiat through multi swap")
		backTokenPrefix := FiatName
		backAmount := "1"
		backAssets := fmt.Sprintf("{\"Assets\":[{\"group\":\"%s\",\"amount\":\"%s\"}]}", backTokenPrefix, backAmount)
		backChannelTo := FiatName
		backSwapBeginArgs := []string{backTokenPrefix, backAssets, backChannelTo, DefaultSwapHash}
		backSignedSwapBeginArgs, err := utils.Sign(userEd25519PrivateKey, userEd25519PublicKey, "cc", "cc", "multiSwapBegin", backSwapBeginArgs)
		assert.NoError(t, err)
		backMultiSwapBegin, err := hlfProxy.Invoke("cc", "multiSwapBegin", backSignedSwapBeginArgs...)
		assert.NoError(t, err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("swapGet txID in fiat channel")
		_, err = hlfProxy.Query("fiat", "multiSwapGet", backMultiSwapBegin.TransactionID)
		assert.NoError(t, err)

		t.NewStep("swapGet txID in cc channel")
		_, err = hlfProxy.Query("cc", "multiSwapGet", backMultiSwapBegin.TransactionID)
		assert.NoError(t, err)

		t.NewStep("After multiSwapBegin need to check balance FIAT token in fiat channel by user address.")
		resp, err = hlfProxy.Query("fiat", "balanceOf", userAddressBase58Check)
		assert.NoError(t, err)
		assert.Equal(t, "\"0\"", string(resp.Payload))

		t.NewStep("After multiSwapBegin need to check allowed balance FIAT token in fiat channel by user address. This balance must change")
		resp, err = hlfProxy.Query("cc", "allowedBalanceOf", userAddressBase58Check, "FIAT")
		assert.NoError(t, err)
		assert.Equal(t, "\"0\"", string(resp.Payload))

		t.NewStep("Complete multi swap process. Invoke multiSwapDone for back FIAT token to 'fiat' channel")
		_, err = hlfProxy.Invoke("fiat", "multiSwapDone", backMultiSwapBegin.TransactionID, DefaultSwapKey)
		assert.NoError(t, err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("After multiSwapDone need to check balance FIAT token in fiat channel by user address. This balance must change")
		resp, err = hlfProxy.Query("fiat", "balanceOf", userAddressBase58Check)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		t.NewStep("After multiSwapDone need to check allowed balance FIAT token in fiat channel by user address")
		resp, err = hlfProxy.Query("cc", "allowedBalanceOf", userAddressBase58Check, "FIAT")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
