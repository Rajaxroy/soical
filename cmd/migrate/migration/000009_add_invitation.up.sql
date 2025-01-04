CREATE TABLE IF NOT EXISTS user_invitations (
    id bytea PRIMARY KEY,
   user_id bigint NOT NULL
);