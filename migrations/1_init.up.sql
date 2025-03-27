CREATE
EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
BEGIN
    IF
NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'usr_type') THEN
CREATE TYPE usr_type AS ENUM('user', 'admin');
END IF;
END$$;

CREATE TABLE IF NOT EXISTS users
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    email TEXT NOT NULL UNIQUE,
    pass_hash VARCHAR
(
    300
) NOT NUll,
    user_type usr_type NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_email ON users (email);