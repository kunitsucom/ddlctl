//nolint:testpackage
package postgres

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	cliz "github.com/kunitsucom/util.go/exp/cli"
	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/fixture"
	ddlctlgo "github.com/kunitsucom/ddlctl/pkg/internal/lang/go"
)

func Test_integrationtest_go_spanner(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cmd := fixture.Cmd()
		args, err := cmd.Parse([]string{
			"ddlctl",
			"--lang=go",
			"--dialect=spanner",
			"--go-column-tag=dbtest",
			"--go-ddl-tag=pgddl",
			"--go-pk-tag=pkey",
			"integrationtest_go_001.source",
			"dummy",
		})
		require.NoError(t, err)

		ctx := cliz.WithContext(context.Background(), cmd)

		{
			_, err := config.Load(ctx)
			require.NoError(t, err)
		}

		ddl, err := ddlctlgo.Parse(ctx, args[1])
		require.NoError(t, err)

		buf := bytes.NewBuffer(nil)

		require.NoError(t, Fprint(buf, ddl))

		golden, err := os.ReadFile("integrationtest_go_001.golden")
		require.NoError(t, err)

		if !assert.Equal(t, string(golden), buf.String()) {
			fmt.Println(buf.String()) //nolint:forbidigo
		}
	})
}
