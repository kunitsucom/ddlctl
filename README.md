# [ddlctl](https://github.com/kunitsucom/ddlctl)

[![license](https://img.shields.io/github/license/kunitsucom/ddlctl)](LICENSE)
[![pkg](https://pkg.go.dev/badge/github.com/kunitsucom/ddlctl)](https://pkg.go.dev/github.com/kunitsucom/ddlctl)
[![goreportcard](https://goreportcard.com/badge/github.com/kunitsucom/ddlctl)](https://goreportcard.com/report/github.com/kunitsucom/ddlctl)
[![workflow](https://github.com/kunitsucom/ddlctl/workflows/go-lint/badge.svg)](https://github.com/kunitsucom/ddlctl/tree/main)
[![workflow](https://github.com/kunitsucom/ddlctl/workflows/go-test/badge.svg)](https://github.com/kunitsucom/ddlctl/tree/main)
[![workflow](https://github.com/kunitsucom/ddlctl/workflows/go-vuln/badge.svg)](https://github.com/kunitsucom/ddlctl/tree/main)
[![codecov](https://codecov.io/gh/kunitsucom/ddlctl/graph/badge.svg?token=8Jtk2bpTe2)](https://codecov.io/gh/kunitsucom/ddlctl)
[![sourcegraph](https://sourcegraph.com/github.com/kunitsucom/ddlctl/-/badge.svg)](https://sourcegraph.com/github.com/kunitsucom/ddlctl)

## Overview

`ddlctl` is a tool for control RDBMS DDL.

## Example

```console
$ # == 1. Prepare your annotated model source code ================================
$ cat <<"EOF" > /tmp/sample.go
package sample

// User is a user model struct.
//
//pgddl:table "users"
//pgddl:constraint UNIQUE ("username")
//pgddl:index "index_users_username" ON "users" ("username")
type User struct {
    UserID   string `db:"user_id"  pgddl:"TEXT NOT NULL" pk:"true"`
    Username string `db:"username" pgddl:"TEXT NOT NULL"`
    Age      int    `db:"age"      pgddl:"INT  NOT NULL"`
}

// Group is a group model struct.
//
//pgddl:table CREATE TABLE IF NOT EXISTS "groups"
//pgddl:index CREATE UNIQUE INDEX "index_groups_group_name" ON "groups" ("group_name")
type Group struct {
    GroupID     string `db:"group_id"    pgddl:"TEXT NOT NULL" pk:"true"`
    GroupName   string `db:"group_name"  pgddl:"TEXT NOT NULL"`
    Description string `db:"description" pgddl:"TEXT NOT NULL"`
}
EOF

$ # == 2. generate DDL ================================
$ ddlctl generate --dialect postgres --column-tag-go db --ddl-tag-go pgddl --pk-tag-go pk --src /tmp/sample.go --dst /tmp/sample.sql
INFO: 2023/11/16 16:10:39 ddlctl.go:44: source: /tmp/sample.go
INFO: 2023/11/16 16:10:39 ddlctl.go:73: destination: /tmp/sample.sql

$ # == 3. Check generated DDL ================================
$ cat /tmp/sample.sql
-- Code generated by ddlctl. DO NOT EDIT.
--

-- source: tmp/sample.go:5
-- User is a user model struct.
--
-- pgddl:table "users"
-- pgddl:constraint UNIQUE ("username")
CREATE TABLE "users" (
    "user_id"  TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "age"      INT  NOT NULL,
    PRIMARY KEY ("user_id"),
    UNIQUE ("username")
);

-- source: tmp/sample.go:7
-- pgddl:index "index_users_username" ON "users" ("username")
CREATE INDEX "index_users_username" ON "users" ("username");

-- source: tmp/sample.go:16
-- Group is a group model struct.
--
-- pgddl:table CREATE TABLE IF NOT EXISTS "groups"
CREATE TABLE IF NOT EXISTS "groups" (
    "group_id"    TEXT NOT NULL,
    "group_name"  TEXT NOT NULL,
    "description" TEXT NOT NULL,
    PRIMARY KEY ("group_id")
);

-- source: tmp/sample.go:17
-- pgddl:index CREATE UNIQUE INDEX "index_groups_group_name" ON "groups" ("group_name")
CREATE UNIQUE INDEX "index_groups_group_name" ON "groups" ("group_name");

```

## Installation

### pre-built binary

```bash
VERSION=v0.0.13

# download
curl -fLROSs https://github.com/kunitsucom/ddlctl/releases/download/${VERSION}/ddlctl_${VERSION}_darwin_arm64.zip

# unzip
unzip -j ddlctl_${VERSION}_darwin_arm64.zip '*/ddlctl'
```

### go install

```bash
go install github.com/kunitsucom/ddlctl/cmd/ddlctl@v0.0.1
```

## Usage

### `ddlctl`

```console
$ ddlctl --help
Usage:
    ddlctl [options]

Description:
    ddlctl is a tool for control RDBMS DDL.

sub commands:
    version: show version
    generate: command "generate" description
    diff: command "diff" description

options:
    --trace (env: DDLCTL_TRACE, default: false)
        trace mode enabled
    --debug (env: DDLCTL_DEBUG, default: false)
        debug mode
    --help (default: false)
        show usage
```

### `ddlctl generate`

```console
$ ddlctl generate --help
Usage:
    ddlctl generate [options] --dialect <DDL dialect> --src <source> --dst <destination>

Description:
    command "ddlctl generate" description

options:
    --lang (env: DDLCTL_LANGUAGE, default: go)
        programming language to generate DDL
    --dialect (env: DDLCTL_DIALECT, default: )
        SQL dialect to generate DDL
    --column-tag-go (env: DDLCTL_COLUMN_TAG_GO, default: db)
        column annotation key for Go struct tag
    --ddl-tag-go (env: DDLCTL_DDL_TAG_GO, default: ddlctl)
        DDL annotation key for Go struct tag
    --pk-tag-go (env: DDLCTL_PK_TAG_GO, default: pk)
        primary key annotation key for Go struct tag
    --src (env: DDLCTL_SOURCE, default: /dev/stdin)
        source file or directory
    --dst (env: DDLCTL_DESTINATION, default: /dev/stdout)
        destination file or directory
    --help (default: false)
        show usage
```

### `ddlctl diff`

```console
$ ddlctl diff --help
Usage:
    ddlctl diff [options] --dialect <DDL dialect> --src <source> --dst <destination>

Description:
    command "ddlctl diff" description

options:
    --lang (env: DDLCTL_LANGUAGE, default: go)
        programming language to generate DDL
    --dialect (env: DDLCTL_DIALECT, default: )
        SQL dialect to generate DDL
    --column-tag-go (env: DDLCTL_COLUMN_TAG_GO, default: db)
        column annotation key for Go struct tag
    --ddl-tag-go (env: DDLCTL_DDL_TAG_GO, default: ddlctl)
        DDL annotation key for Go struct tag
    --pk-tag-go (env: DDLCTL_PK_TAG_GO, default: pk)
        primary key annotation key for Go struct tag
    --help (default: false)
        show usage
```

## TODO

- dialect
  - `generate` subcommand
    - [x] Support `mysql`
    - [x] Support `postgres`
    - [x] Support `cockroachdb`
    - [x] Support `spanner`
    - [ ] Support `sqlite3`
  - `diff` subcommand
    - [ ] Support `mysql`
    - [ ] Support `postgres`
    - [x] Support `cockroachdb`
    - [ ] Support `spanner`
    - [ ] Support `sqlite3`
- lang
  - [x] Support `go`