package sample

// User is a user model struct.
//
//pgddl:table      public.users
//pgddl:constraint UNIQUE ("username")
//pgddl:index      "index_users_username" ON public.users ("username")
type User struct {
	UserID   string `db:"user_id"  pgddl:"TEXT NOT NULL" pk:"true"`
	Username string `db:"username" pgddl:"TEXT NOT NULL"`
	Age      int    `db:"age"      pgddl:"INT  NOT NULL"`
}

// Group is a group model struct.
//
//pgddl:table CREATE TABLE IF NOT EXISTS public.groups
//pgddl:index CREATE UNIQUE INDEX "index_groups_group_name" ON public.groups ("group_name")
type Group struct {
	GroupID     string `db:"group_id"    pgddl:"TEXT NOT NULL" pk:"true"`
	GroupName   string `db:"group_name"  pgddl:"TEXT NOT NULL"`
	Description string `db:"description" pgddl:"TEXT NOT NULL"`
}
