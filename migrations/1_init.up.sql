CREATE EXTENSION IF NOT EXISTS "pgcrypto";

create type usr_type as ENUM('customer', 'seller');

CREATE TABLE IF NOT EXISTS users
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    pass_hash VARCHAR(300) NOT NUll,
    user_type usr_type NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_email ON users (email);