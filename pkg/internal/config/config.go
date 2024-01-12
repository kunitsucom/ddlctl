package config

import (
	"context"
	"encoding/json"
	"sync"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/pkg/internal/logs"
)

// Use a structure so that settings can be backed up.
//
//nolint:tagliatelle
type config struct {
	Version     bool   `json:"version"`
	Trace       bool   `json:"trace"`
	Debug       bool   `json:"debug"`
	Language    string `json:"language"`
	Dialect     string `json:"dialect"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	AutoApprove bool   `json:"auto_approve"`
	// Golang
	ColumnTagGo string `json:"column_tag_go"`
	DDLTagGo    string `json:"ddl_tag_go"`
	PKTagGo     string `json:"pk_tag_go"`
}

//nolint:gochecknoglobals
var (
	globalConfig   *config
	globalConfigMu sync.RWMutex
)

func MustLoad(ctx context.Context) (rollback func()) {
	rollback, err := Load(ctx)
	if err != nil {
		err = errorz.Errorf("Load: %w", err)
		panic(err)
	}
	return rollback
}

func Load(ctx context.Context) (rollback func(), err error) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	backup := globalConfig

	cfg, err := load(ctx)
	if err != nil {
		return nil, errorz.Errorf("load: %w", err)
	}

	globalConfig = cfg

	rollback = func() {
		globalConfigMu.Lock()
		defer globalConfigMu.Unlock()
		globalConfig = backup
	}

	return rollback, nil
}

// MEMO: Since there is a possibility of returning some kind of error in the future, the signature is made to return an error.
//
//nolint:funlen
func load(ctx context.Context) (cfg *config, err error) { //nolint:unparam
	cmd, err := cliz.FromContext(ctx)
	if err != nil {
		return nil, errorz.Errorf("cliz.FromContext: %w", err)
	}

	c := &config{
		Trace:       loadTrace(ctx, cmd),
		Debug:       loadDebug(ctx, cmd),
		Language:    loadLanguage(ctx, cmd),
		Dialect:     loadDialect(ctx, cmd),
		Source:      loadSource(ctx, cmd),
		Destination: loadDestination(ctx, cmd),
		AutoApprove: loadAutoApprove(ctx, cmd),
		ColumnTagGo: loadColumnTagGo(ctx, cmd),
		DDLTagGo:    loadDDLTagGo(ctx, cmd),
		PKTagGo:     loadPKTagGo(ctx, cmd),
	}

	if c.Debug {
		logs.Debug = logs.NewDebug()
		logs.Debug.Print("debug mode enabled")
	}
	if c.Trace {
		logs.Trace = logs.NewTrace()
		logs.Debug = logs.NewDebug()
		logs.Trace.Print("trace mode enabled")
	}

	if err := json.NewEncoder(logs.Debug).Encode(c); err != nil {
		logs.Debug.Printf("config: %#v", c)
	}

	return c, nil
}
