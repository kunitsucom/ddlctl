package ddlctl

import (
	"context"
	"errors"
	"fmt"
	"os"

	cliz "github.com/kunitsucom/util.go/exp/cli"
	"github.com/kunitsucom/util.go/version"

	apperr "github.com/kunitsucom/ddlctl/pkg/apperr"
	"github.com/kunitsucom/ddlctl/pkg/ddlctl/apply"
	"github.com/kunitsucom/ddlctl/pkg/ddlctl/diff"
	"github.com/kunitsucom/ddlctl/pkg/ddlctl/generate"
	"github.com/kunitsucom/ddlctl/pkg/ddlctl/show"
	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

//nolint:gochecknoglobals
var (
	optLanguage = &cliz.StringOption{
		Name:        consts.OptionLanguage,
		Environment: consts.EnvKeyLanguage,
		Description: "programming language to generate DDL",
		Default:     cliz.Default("go"),
	}
	optDialect = &cliz.StringOption{
		Name:        consts.OptionDialect,
		Environment: consts.EnvKeyDialect,
		Description: "SQL dialect to generate DDL",
		Default:     cliz.Default(""),
	}
	opts = []cliz.Option{
		optLanguage,
		optDialect,
		// Golang
		&cliz.StringOption{
			Name:        consts.OptionGoColumnTag,
			Environment: consts.EnvKeyGoColumnTag,
			Description: "column annotation key for Go struct tag",
			Default:     cliz.Default("db"),
		},
		&cliz.StringOption{
			Name:        consts.OptionGoDDLTag,
			Environment: consts.EnvKeyGoDDLTag,
			Description: "DDL annotation key for Go struct tag",
			Default:     cliz.Default("ddlctl"),
		},
		&cliz.StringOption{
			Name:        consts.OptionGoPKTag,
			Environment: consts.EnvKeyGoPKTag,
			Description: "primary key annotation key for Go struct tag",
			Default:     cliz.Default("pk"),
		},
	}
)

//nolint:cyclop,funlen
func DDLCtl(ctx context.Context) error {
	cmd := cliz.Command{
		Name:        "ddlctl",
		Usage:       "ddlctl [options]",
		Description: "ddlctl is a tool for control RDBMS DDL.",
		SubCommands: []*cliz.Command{
			{
				Name:        "version",
				Usage:       "ddlctl version",
				Description: "show version",
				RunFunc: func(_ context.Context, _ []string) error {
					fmt.Printf("version: %s\n", version.Version())           //nolint:forbidigo
					fmt.Printf("revision: %s\n", version.Revision())         //nolint:forbidigo
					fmt.Printf("build branch: %s\n", version.Branch())       //nolint:forbidigo
					fmt.Printf("build timestamp: %s\n", version.Timestamp()) //nolint:forbidigo
					return nil
				},
			},
			{
				Name:        "generate",
				Short:       "gen",
				Usage:       "ddlctl generate [options] --dialect <DDL dialect> <source> <destination>",
				Description: "generate DDL from source (file or directory) to destination (file or directory).",
				Options:     opts,
				RunFunc:     generate.Command,
			},
			{
				Name:        "show",
				Usage:       "ddlctl show --dialect <DDL dialect> <DSN>",
				Description: "show DDL from DSN like `SHOW CREATE TABLE`.",
				Options:     []cliz.Option{optDialect},
				RunFunc:     show.Command,
			},
			{
				Name:        "diff",
				Usage:       "ddlctl diff [options] --dialect <DDL dialect> <before DDL source> <after DDL source>",
				Description: "diff DDL from <before DDL source> to <after DDL source>.",
				Options:     opts,
				RunFunc:     diff.Command,
			},
			{
				Name:        "apply",
				Usage:       "ddlctl apply [options] --dialect <DDL dialect> <DSN to apply> <DDL source>",
				Description: "apply DDL from <DDL source> to <DSN to apply>.",
				Options: append(opts,
					&cliz.BoolOption{
						Name:        consts.OptionAutoApprove,
						Environment: consts.EnvKeyAutoApprove,
						Description: "auto approve",
						Default:     cliz.Default(false),
					},
				),
				RunFunc: apply.Command,
			},
		},
		Options: []cliz.Option{
			&cliz.BoolOption{
				Name:        consts.OptionTrace,
				Environment: consts.EnvKeyTrace,
				Description: "trace mode enabled",
				Default:     cliz.Default(false),
			},
			&cliz.BoolOption{
				Name:        consts.OptionDebug,
				Environment: consts.EnvKeyDebug,
				Description: "debug mode",
				Default:     cliz.Default(false),
			},
		},
		RunFunc: func(ctx context.Context, _ []string) error {
			cmd, err := cliz.FromContext(ctx)
			if err != nil {
				return apperr.Errorf("cliz.FromContext: %w", err)
			}

			cmd.ShowUsage()
			return cliz.ErrHelp
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		if errors.Is(err, cliz.ErrHelp) {
			return nil
		}

		return apperr.Errorf("cmd.Run: %w", err)
	}

	return nil
}
