package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

func loadDDLTagGo(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(consts.OptionDDLTagGo)
	return v
}

func DDLTagGo() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.DDLTagGo
}
