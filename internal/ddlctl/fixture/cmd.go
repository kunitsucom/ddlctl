package fixture

import (
	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/ddlctl/internal/consts"
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
