CREATE TABLE IF NOT EXISTS public.test_groups (
    group_id   UUID NOT NULL,
    group_name TEXT NOT NULL,
    PRIMARY KEY (group_id)
);
CREATE TABLE IF NOT EXISTS public.test_users (
    user_id  UUID NOT NULL,
    name     TEXT NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (user_id)
);
CREATE INDEX IF NOT EXISTS test_users_idx_on_group_id ON public.test_users (group_id);
