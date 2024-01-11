package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

func loadLanguage(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(consts.OptionLanguage)
	return v
}

func Language() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.Language
}
