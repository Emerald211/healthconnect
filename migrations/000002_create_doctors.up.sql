CREATE TABLE doctors (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(100)    NOT NULL,
    email             VARCHAR(255)    NOT NULL UNIQUE,
    phone             VARCHAR(20)     NOT NULL,
    password          TEXT            NOT NULL,
    specialty         VARCHAR(100)    NOT NULL,
    license_number    VARCHAR(50)     NOT NULL UNIQUE,
    years_experience  INT             NOT NULL DEFAULT 0,
    consultation_fee  DECIMAL(10,2)   NOT NULL DEFAULT 0.00,
    bio               TEXT,
    is_active         BOOLEAN         NOT NULL DEFAULT true,
    is_verified       BOOLEAN         NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_doctors_email     ON doctors(email);
CREATE INDEX idx_doctors_specialty ON doctors(specialty);