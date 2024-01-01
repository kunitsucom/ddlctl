package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/internal/consts"
)

func loadPKTagGo(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(consts.OptionPKTagGo)
	return v
}

func PKTagGo() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.PKTagGo
}
