package ddlctl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/ddlctl/internal/config"
	"github.com/kunitsucom/ddlctl/internal/consts"
	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
)

//nolint:cyclop,funlen
func Apply(ctx context.Context, args []string) error {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	if len(args) != 2 {
		return errorz.Errorf("args=%v: %w", args, apperr.ErrTwoArgumentsRequired)
	}

	dsn := args[0]
	ddlSrc := args[1]

	left, right, err := resolve(ctx, config.Dialect(), dsn, ddlSrc)
	if err != nil {
		return errorz.Errorf("resolve: %w", err)
	}

	buf := new(strings.Builder)
	if err := diff(buf, left, right); err != nil {
		return errorz.Errorf("diff: %w", err)
	}

	msg := `
ddlctl will exec the following DDL queries:

-- 8< --

` + buf.String() + `

-- >8 --

Do you want to apply these DDL queries?
  ddlctl will exec the DDL queries described above.
  Only 'yes' will be accepted to approve.

Enter a value: `

	if _, err := os.Stdout.WriteString(msg); err != nil {
		return errorz.Errorf("os.Stdout.WriteString: %w", err)
	}

	if config.AutoApprove() {
		if _, err := os.Stdout.WriteString(fmt.Sprintf("yes (via --%s option)\n", consts.OptionAutoApprove)); err != nil {
			return errorz.Errorf("os.Stdout.WriteString: %w", err)
		}
	} else {
		if err := prompt(); err != nil {
			return errorz.Errorf("prompt: %w", err)
		}
	}

	os.Stdout.WriteString("\nexecuting...\n")

	db, err := sqlz.OpenContext(ctx, _postgres, dsn)
	if err != nil {
		return errorz.Errorf("sqlz.OpenContext: %w", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, buf.String()); err != nil {
		return errorz.Errorf("db.ExecContext: %w", err)
	}

	os.Stdout.WriteString("done\n")

	return nil
}

func prompt() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := scanner.Text()

	switch userInput {
	case "yes":
		return nil
	default:
		return errorz.Errorf("userInput=%s: %w", userInput, apperr.ErrUserCanceled)
	}
}
