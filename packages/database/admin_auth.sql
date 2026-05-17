-- Admin authentication schema for UrbanMemory
-- PostgreSQL / PostGIS compatible

CREATE TABLE IF NOT EXISTS admins (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'pending'
        CHECK (role IN ('super_admin', 'approved', 'pending')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admins_email ON admins (email);
CREATE INDEX IF NOT EXISTS idx_admins_role ON admins (role);

CREATE TABLE IF NOT EXISTS otp_verifications (
    id BIGSERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    otp_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'notary',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_otp_verifications_admin_id ON otp_verifications (admin_id);
CREATE INDEX IF NOT EXISTS idx_otp_verifications_expires_at ON otp_verifications (expires_at);
CREATE INDEX IF NOT EXISTS idx_otp_verifications_used ON otp_verifications (used);
CREATE INDEX IF NOT EXISTS idx_otp_verifications_purpose ON otp_verifications (purpose);
