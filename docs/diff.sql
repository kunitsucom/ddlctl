-- Code generated by ddlctl. DO NOT EDIT.
--

-- source: docs/sample.go:5
-- User is a user model struct.
--
-- pgddl:table      public.users
-- pgddl:constraint UNIQUE ("username")
CREATE TABLE public.users (
    "user_id"     TEXT NOT NULL,
    "username"    TEXT NOT NULL,
    "age"         INT  NOT NULL,
    "description" TEXT NOT NULL,
    PRIMARY KEY ("user_id"),
    UNIQUE ("username")
);

-- source: docs/sample.go:7
-- pgddl:index      "index_users_username" ON public.users ("username")
CREATE INDEX "index_users_username" ON public.users ("username");

-- source: docs/sample.go:16
-- Group is a group model struct.
--
-- pgddl:table CREATE TABLE IF NOT EXISTS public.groups
CREATE TABLE IF NOT EXISTS public.groups (
    "group_id"    TEXT NOT NULL,
    "group_name"  TEXT NOT NULL,
    "description" TEXT NOT NULL,
    PRIMARY KEY ("group_id")
);

-- source: docs/sample.go:17
-- pgddl:index CREATE UNIQUE INDEX "index_groups_group_name" ON public.groups ("group_name")
CREATE UNIQUE INDEX "index_groups_group_name" ON public.groups ("group_name");
