-- Create patients table
-- This stores everyone who registers as a patient on HealthConnect

CREATE TABLE patients (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100)        NOT NULL,
    email       VARCHAR(255)        NOT NULL UNIQUE,
    phone       VARCHAR(20)         NOT NULL,
    password    TEXT                NOT NULL,
    date_of_birth DATE              NOT NULL,
    gender      VARCHAR(10)         NOT NULL,
    address     TEXT,
    is_active   BOOLEAN             NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

-- Index on email — we'll query patients by email often (login)
-- Without this index, every login scans the entire table
CREATE INDEX idx_patients_email ON patients(email);

-- Index on phone — used for SMS lookup
CREATE INDEX idx_patients_phone ON patients(phone); 