package fixture

import (
	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

func Cmd() *cliz.Command {
	return &cliz.Command{
		Options: []cliz.Option{
			&cliz.StringOption{
				Name:        consts.OptionLanguage,
				Environment: consts.EnvKeyLanguage,
				Description: "programming language to generate DDL",
				Default:     cliz.Default("go"),
			},
			&cliz.StringOption{
				Name:        consts.OptionDialect,
				Environment: consts.EnvKeyDialect,
				Description: "SQL dialect to generate DDL",
				Default:     cliz.Default(""),
			},
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
		},
	}
}
