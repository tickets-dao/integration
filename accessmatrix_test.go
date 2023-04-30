package integration

import (
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/golang/protobuf/proto"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
	pb "github.com/tickets-dao/integration/proto"
	"github.com/tickets-dao/integration/utils"
)

const acl = "acl"
const issuer = "issuer"

const (
	addUserFn       = "addUser"
	addRightsFn     = "addRights"
	removeRightsFn  = "removeRights"
	getAccOpRightFn = "getAccountOperationRight"
)

// TestAclRights creates user in acl chaincode, adds rights, removes rights, checks all operations is done
func TestAclRights(t *testing.T) {
	runner.Run(t, "Add and remove rights", func(t provider.T) {
		t.Tags("positive", "acl")

		t.NewStep("Prepare service to connect with API hlf proxy service")
		hlfProxy := utils.NewHlfProxyService(os.Getenv(utils.EnvHlfProxyURL), os.Getenv(utils.EnvHlfProxyAuthToken))

		t.NewStep("Generate private key for user")
		_, userFromEd25519PublicKey, err := utils.GeneratePrivateAndPublicKey()
		t.Assert().NoError(err)

		t.NewStep("Get publicKeyBase58")
		publicKeyBase58 := base58.Encode(userFromEd25519PublicKey)

		t.NewStep("Get user address")
		userAddress, err := utils.GetAddressByPublicKey(userFromEd25519PublicKey)
		assert.NoError(t, err)

		t.NewStep("Invoke chaincode acl with method addUser, and create user")
		_, err = hlfProxy.Invoke("acl", "addUser", publicKeyBase58, "test", "testuser", "true")
		t.Assert().NoError(err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("Query chaincode acl with method checkKeys")
		_, err = hlfProxy.Query("acl", "checkKeys", publicKeyBase58)
		t.Assert().NoError(err)

		t.NewStep("Invoke chaincode `" + acl + "` with method `" + addUserFn + "`, and create user")
		_, err = hlfProxy.Invoke(acl, addUserFn, userAddress, "test", "testuser", "true")
		t.Assert().NoError(err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("Query chaincode `acl` with method checkKeys")
		_, err = hlfProxy.Query(acl, "checkKeys", userAddress)
		t.Assert().NoError(err)

		const testOperation = "testOperation"

		t.NewStep("Invoke chaincode `" + acl + "` with method `" + addRightsFn + "` and grant right")
		_, err = hlfProxy.Invoke(acl, addRightsFn, acl, acl, issuer, testOperation, userAddress)
		t.Assert().NoError(err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("Query chaincode `" + acl + "` with method `" + getAccOpRightFn + "` rights is set")
		rsp, err := hlfProxy.Query(acl, getAccOpRightFn, acl, acl, issuer, testOperation, userAddress)
		t.Assert().NoError(err)
		var haveRight pb.HaveRight
		err = proto.Unmarshal(rsp.Payload, &haveRight)
		t.Assert().NoError(err)
		t.Assert().Equal(true, haveRight.HaveRight)

		t.NewStep("Invoke chaincode `" + acl + "` with method `" + removeRightsFn + "` and remove right")
		_, err = hlfProxy.Invoke(acl, removeRightsFn, acl, acl, issuer, testOperation, userAddress)
		t.Assert().NoError(err)
		time.Sleep(utils.BatchTransactionTimeout)

		t.NewStep("Query chaincode `" + acl + "` with method `" + getAccOpRightFn + "` rights is not set")
		rsp2, err := hlfProxy.Query(acl, getAccOpRightFn, acl, acl, issuer, testOperation, userAddress)
		t.Assert().NoError(err)
		var r pb.HaveRight
		err = proto.Unmarshal(rsp2.Payload, &r)
		t.Assert().NoError(err)
		t.Assert().Equal(false, r.HaveRight)
	})
}
