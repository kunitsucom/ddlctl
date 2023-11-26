package main

import (
	"context"
	"log"

	"github.com/kunitsucom/ddlctl/pkg/ddlctl"
)

func main() {
	ctx := context.Background()

	if err := ddlctl.DDLCtl(ctx); err != nil {
		log.Fatalf("ddlctl: %+v", err)
	}
}
