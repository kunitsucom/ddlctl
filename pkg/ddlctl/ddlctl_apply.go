package ddlctl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	stringz "github.com/kunitsucom/util.go/strings"

	apperr "github.com/kunitsucom/ddlctl/pkg/apperr"
	"github.com/kunitsucom/ddlctl/pkg/ddl"
	crdbddl "github.com/kunitsucom/ddlctl/pkg/ddl/cockroachdb"
	spanddl "github.com/kunitsucom/ddlctl/pkg/ddl/spanner"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
)

//nolint:cyclop,funlen,gocognit
func Apply(ctx context.Context, args []string) (err error) {
	if _, err := config.Load(ctx); err != nil {
		return apperr.Errorf("config.Load: %w", err)
	}

	if len(args) != 2 {
		return apperr.Errorf("args=%v: %w", args, apperr.ErrTwoArgumentsRequired)
	}

	dsn, ddlSrc := args[0], args[1]
	dialect := config.Dialect()

	left, err := resolve(ctx, dialect, dsn)
	if err != nil {
		return apperr.Errorf("resolve: %w", err)
	}

	right, err := resolve(ctx, dialect, ddlSrc)
	if err != nil {
		return apperr.Errorf("resolve: %w", err)
	}

	buf := new(strings.Builder)
	if err := DiffDDL(buf, dialect, left, right); err != nil {
		if errors.Is(err, ddl.ErrNoDifference) {
			_, _ = fmt.Fprintln(os.Stdout, ddl.ErrNoDifference.Error())
			return nil
		}
		return apperr.Errorf("diff: %w", err)
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
		return apperr.Errorf("os.Stdout.WriteString: %w", err)
	}

	if config.AutoApprove() {
		if _, err := os.Stdout.WriteString(fmt.Sprintf("yes (via --%s option)\n", consts.OptionAutoApprove)); err != nil {
			return apperr.Errorf("os.Stdout.WriteString: %w", err)
		}
	} else {
		if err := prompt(); err != nil {
			return apperr.Errorf("prompt: %w", err)
		}
	}

	os.Stdout.WriteString("\nexecuting...\n")

	driverName := func() string {
		switch dialect {
		case crdbddl.Dialect:
			return crdbddl.DriverName
		default:
			return dialect
		}
	}()

	db, err := sqlz.OpenContext(ctx, driverName, dsn)
	if err != nil {
		return apperr.Errorf("sqlz.OpenContext: %w", err)
	}
	defer func() {
		if cerr := db.Close(); err == nil && cerr != nil {
			err = apperr.Errorf("db.Close: %w", cerr)
		}
	}()

	switch {
	case driverName == spanddl.DriverName:
		conn, err := db.Conn(ctx)
		if err != nil {
			return apperr.Errorf("db.Conn: %w", err)
		}
		defer func() {
			if cerr := conn.Close(); err == nil && cerr != nil {
				err = apperr.Errorf("conn.Close: %w", cerr)
			}
		}()
		if _, err := conn.ExecContext(ctx, "START BATCH DDL"); err != nil {
			return apperr.Errorf("conn.ExecContext: %w", err)
		}

		commentTrimmedDDL := stringz.ReadLine(q, "\n", stringz.ReadLineFuncRemoveCommentLine("--"))
		for _, q := range strings.Split(commentTrimmedDDL, ";\n") {
			if len(q) == 0 {
				continue
			}
			if _, err := conn.ExecContext(ctx, q); err != nil {
				return apperr.Errorf("conn.ExecContext: %w", err)
			}
		}

		if _, err := conn.ExecContext(ctx, "RUN BATCH"); err != nil {
			return apperr.Errorf("conn.ExecContext: %w", err)
		}
	default:
		if _, err := db.ExecContext(ctx, q); err != nil {
			return apperr.Errorf("db.ExecContext: %w", err)
		}
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
		return apperr.Errorf("input=%s: %w", input, apperr.ErrCanceled)
	}
}
