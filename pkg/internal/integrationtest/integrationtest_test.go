package integrationtest_test

import (
	"context"
	"os"
	"testing"

	cliz "github.com/kunitsucom/util.go/exp/cli"
	testingz "github.com/kunitsucom/util.go/testing"
	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/ddlctl/pkg/ddlctl"
	"github.com/kunitsucom/ddlctl/pkg/internal/fixture"
)

//nolint:paralleltest
func Test_ddlctl_diff(t *testing.T) {
	t.Run("success,go,postgres", func(t *testing.T) {
		cmd := fixture.Cmd()
		args, err := cmd.Parse([]string{
			"--lang=go",
			"--dialect=postgres",
			"postgres_before.sql",
			"postgres_after.sql",
		})
		require.NoError(t, err)
		ctx := cliz.WithContext(context.Background(), cmd)

		backup := os.Stdout
		t.Cleanup(func() { os.Stdout = backup })

		w, closeFunc, err := testingz.NewFileWriter(t)
		require.NoError(t, err)

		os.Stdout = w
		{
			err := ddlctl.Diff(ctx, args)
			require.NoError(t, err)
		}
		result := closeFunc()

		const expected = `-- -
-- +description TEXT NOT NULL
ALTER TABLE public.test_groups ADD COLUMN description TEXT NOT NULL;
-- -name TEXT NOT NULL
-- +
ALTER TABLE public.test_users DROP COLUMN name;
-- -
-- +username TEXT NOT NULL
ALTER TABLE public.test_users ADD COLUMN username TEXT NOT NULL;
`

		actual := result.String()

		assert.Equal(t, expected, actual)
	})
}
