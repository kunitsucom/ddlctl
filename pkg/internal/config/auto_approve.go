package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

func loadAutoApprove(_ context.Context, cmd *cliz.Command) bool {
	v, _ := cmd.GetOptionBool(consts.OptionAutoApprove)
	return v
}

func AutoApprove() bool {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.AutoApprove
}
