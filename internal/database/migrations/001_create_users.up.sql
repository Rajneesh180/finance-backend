CREATE TABLE IF NOT EXISTS users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'viewer'
                CHECK (role IN ('admin', 'analyst', 'viewer')),
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role  ON users(role)  WHERE deleted_at IS NULL;
