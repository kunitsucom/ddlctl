CREATE TABLE public.test_groups (
    group_id    UUID NOT NULL,
    group_name  TEXT NOT NULL,
    description TEXT NOT NULL,
    PRIMARY KEY (group_id)
);
CREATE TABLE public.test_users (
    user_id  UUID NOT NULL,
    username TEXT NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (user_id)
);
CREATE INDEX test_users_idx_on_group_id ON public.test_users (group_id);
