package ddlctl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"

	crdbddl "github.com/kunitsucom/ddlctl/pkg/ddl/cockroachdb"
	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

//nolint:cyclop,funlen
func Apply(ctx context.Context, args []string) (err error) {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	if len(args) != 2 {
		return errorz.Errorf("args=%v: %w", args, apperr.ErrTwoArgumentsRequired)
	}

	dsn, ddlSrc := args[0], args[1]

	left, err := resolve(ctx, config.Dialect(), dsn)
	if err != nil {
		return errorz.Errorf("resolve: %w", err)
	}

	right, err := resolve(ctx, config.Dialect(), ddlSrc)
	if err != nil {
		return errorz.Errorf("resolve: %w", err)
	}

	buf := new(strings.Builder)
	if err := DiffDDL(buf, config.Dialect(), left, right); err != nil {
		return errorz.Errorf("diff: %w", err)
	}
	q := buf.String()

	msg := `
ddlctl will exec the following DDL queries:

-- 8< --

` + q + `

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

	driverName := func() string {
		switch dialect := config.Dialect(); dialect {
		case crdbddl.Dialect:
			return crdbddl.DriverName
		default:
			return dialect
		}
	}()

	db, err := sqlz.OpenContext(ctx, driverName, dsn)
	if err != nil {
		return errorz.Errorf("sqlz.OpenContext: %w", err)
	}
	defer func() {
		if cerr := db.Close(); err == nil && cerr != nil {
			err = errorz.Errorf("db.Close: %w", cerr)
		}
	}()

	if _, err := db.ExecContext(ctx, q); err != nil {
		return errorz.Errorf("db.ExecContext: %w", err)
	}

	os.Stdout.WriteString("done\n")

	return nil
}

func prompt() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	switch input {
	case "yes":
		return nil
	default:
		return errorz.Errorf("input=%s: %w", input, apperr.ErrCanceled)
	}
}
