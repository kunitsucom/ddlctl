package apply

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"
	"github.com/kunitsucom/util.go/retry"
	stringz "github.com/kunitsucom/util.go/strings"

	apperr "github.com/kunitsucom/ddlctl/pkg/apperr"
	"github.com/kunitsucom/ddlctl/pkg/ddl"
	ddlcrdb "github.com/kunitsucom/ddlctl/pkg/ddl/cockroachdb"
	ddlmysql "github.com/kunitsucom/ddlctl/pkg/ddl/mysql"
	ddlspanner "github.com/kunitsucom/ddlctl/pkg/ddl/spanner"
	"github.com/kunitsucom/ddlctl/pkg/ddlctl/diff"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/consts"
	"github.com/kunitsucom/ddlctl/pkg/logs"
)

//nolint:cyclop,funlen,gocognit,gocyclo
func Command(ctx context.Context, args []string) (err error) {
	if _, err := config.Load(ctx); err != nil {
		return apperr.Errorf("config.Load: %w", err)
	}

	const beforeAndAfterForDiff = 2
	if len(args) != beforeAndAfterForDiff {
		return apperr.Errorf("args=%v: %w", args, apperr.ErrTwoArgumentsRequired)
	}

	dialect := config.Dialect()
	language := config.Language()
	leftArg, rightArg := args[0], args[1]

	buf := new(strings.Builder)
	if err := diff.Diff(ctx, buf, dialect, language, leftArg, rightArg); err != nil {
		if errors.Is(err, ddl.ErrNoDifference) {
			_, _ = fmt.Fprintln(os.Stdout, ddl.ErrNoDifference.Error())
			return nil
		}
		return apperr.Errorf("diff: %w", err)
	}
	ddlStr := buf.String()

	msg := `
ddlctl will exec the following DDL queries:

-- 8< --

` + ddlStr + `

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

	if err := Apply(ctx, dialect, leftArg, ddlStr); err != nil {
		return apperr.Errorf("Apply: %w", err)
	}

	os.Stdout.WriteString("done\n")

	return nil
}

//nolint:cyclop,funlen,gocognit,gocyclo
func Apply(ctx context.Context, dialect, dsn, ddlStr string) error {
	driverName := func() string {
		switch dialect {
		case ddlcrdb.Dialect:
			return ddlcrdb.DriverName
		default:
			return dialect
		}
	}()

	db, err := sqlz.OpenContext(ctx, driverName, dsn)
	if err != nil {
		return apperr.Errorf("sqlz.OpenContext: %w", err)
	}
	defer func() {
		if err2 := db.Close(); err2 != nil && err == nil {
			err = apperr.Errorf("db.Close: %w", err2)
		}
	}()

	switch driverName {
	case ddlmysql.DriverName:
		ddls := strings.Split(ddlStr, ";\n")
		retryer := retry.New(retry.NewConfig(time.Second, time.Second, retry.WithMaxRetries(len(ddls))))
		if err := retryer.Do(ctx, func(ctx context.Context) error {
			var outerErr error
			for _, q := range ddls {
				if len(q) == 0 {
					// skip empty query
					continue
				}
				if _, err := db.ExecContext(ctx, q); err != nil {
					// not error, not log, go next ddl
					if errorz.Contains(err, "already exists") || errorz.Contains(err, "Duplicate column name") {
						continue
					}
					outerErr = err
					// error, but not log, go next ddl
					if errorz.Contains(err, "Cannot add foreign key constraint") {
						continue
					}
					// error and log, go next ddl
					logs.Warn.Printf("db.ExecContext: %v", err)
				}
			}
			if outerErr != nil {
				return apperr.Errorf("db.ExecContext: %w", err)
			}
			return nil
		}); err != nil {
			return apperr.Errorf("retry.Do: %w", err)
		}

	case ddlspanner.DriverName:
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

		commentTrimmedDDL := stringz.ReadLine(ddlStr, "\n", stringz.ReadLineFuncRemoveCommentLine("--"))
		for _, q := range strings.Split(commentTrimmedDDL, ";\n") {
			if len(q) == 0 {
				// skip empty query
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
		if _, err := db.ExecContext(ctx, ddlStr); err != nil {
			return apperr.Errorf("db.ExecContext: %w", err)
		}
	}

	return nil
}

func prompt() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	switch input := scanner.Text(); input {
	case "yes":
		return nil
	default:
		return apperr.Errorf("input=%q: %w", input, apperr.ErrCanceled)
	}
}
