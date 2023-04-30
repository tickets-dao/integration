package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/tickets-dao/integration/utils"
)

func TestMetadata(t *testing.T) {
	ccs := []string{`cc`, `fiat`, `industrial`}

	runner.Run(t, "Check metadata in "+strings.Join(ccs, ", "), func(t provider.T) {
		ctx := context.Background()
		t.Severity(allure.BLOCKER)
		t.Description("Acceptance of existence of method `metadata` in " + strings.Join(ccs, ", ") + " chaincodes")
		t.Tags("smoke", "positive", "metadata")
		for _, cc := range ccs {
			t.WithNewAsyncStep("Get metadata from chaincode `"+cc+"`", func(sCtx provider.StepCtx) {
				_, err := utils.Query(ctx, os.Getenv(utils.EnvHlfProxyURL),
					os.Getenv(utils.EnvHlfProxyAuthToken), cc, "metadata")
				sCtx.Assert().NoError(err)
			})
		}
	})
}
