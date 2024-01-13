package ddlctl

import (
	"context"
	"errors"
	"fmt"
	"os"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"
	"github.com/kunitsucom/util.go/version"

	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

const (
	_spanner = "spanner" // TODO: remove after spanner ddl diff implemented
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
			Name:        consts.OptionColumnTagGo,
			Environment: consts.EnvKeyColumnTagGo,
			Description: "column annotation key for Go struct tag",
			Default:     cliz.Default("db"),
		},
		&cliz.StringOption{
			Name:        consts.OptionDDLTagGo,
			Environment: consts.EnvKeyDDLTagGo,
			Description: "DDL annotation key for Go struct tag",
			Default:     cliz.Default("ddlctl"),
		},
		&cliz.StringOption{
			Name:        consts.OptionPKTagGo,
			Environment: consts.EnvKeyPKTagGo,
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
		Description: `ddlctl is a tool for control RDBMS DDL.`,
		SubCommands: []*cliz.Command{
			{
				Name:        "version",
				Usage:       "ddlctl version",
				Description: "show version",
				RunFunc: func(ctx context.Context, args []string) error {
					fmt.Printf("version: %s\n", version.Version())           //nolint:forbidigo
					fmt.Printf("revision: %s\n", version.Revision())         //nolint:forbidigo
					fmt.Printf("build branch: %s\n", version.Branch())       //nolint:forbidigo
					fmt.Printf("build timestamp: %s\n", version.Timestamp()) //nolint:forbidigo
					return nil
				},
			},
			{
				Name:  "generate",
				Short: "gen",
				Usage: "ddlctl generate [options] --dialect <DDL dialect> --src <source> --dst <destination>",
				Options: append(opts,
					&cliz.StringOption{
						Name:        consts.OptionSource,
						Environment: consts.EnvKeySource,
						Description: "source file or directory",
						Default:     cliz.Default("/dev/stdin"),
					},
					&cliz.StringOption{
						Name:        consts.OptionDestination,
						Environment: consts.EnvKeyDestination,
						Description: "destination file or directory",
						Default:     cliz.Default("/dev/stdout"),
					},
				),
				RunFunc: Generate,
			},
			{
				Name:    "show",
				Usage:   "ddlctl show --dialect <DDL dialect> <DSN>",
				Options: []cliz.Option{optDialect},
				RunFunc: Show,
			},
			{
				Name:    "diff",
				Usage:   "ddlctl diff [options] --dialect <DDL dialect> <DDL source before> <DDL source after>",
				Options: opts,
				RunFunc: Diff,
			},
			{
				Name:  "apply",
				Usage: "ddlctl apply [options] --dialect <DDL dialect> <DSN to apply> <DDL source>",
				Options: append(opts,
					&cliz.BoolOption{
						Name:        consts.OptionAutoApprove,
						Environment: consts.EnvKeyAutoApprove,
						Description: "auto approve",
						Default:     cliz.Default(false),
					},
				),
				RunFunc: Apply,
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
		RunFunc: func(ctx context.Context, args []string) error {
			cmd, err := cliz.FromContext(ctx)
			if err != nil {
				return errorz.Errorf("cliz.FromContext: %w", err)
			}

			cmd.ShowUsage()
			return cliz.ErrHelp
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		if errors.Is(err, cliz.ErrHelp) {
			return nil
		}
		return errorz.Errorf("cmd.Run: %w", err)
	}

	return nil
}
