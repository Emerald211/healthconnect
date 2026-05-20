-- Doctor weekly availability schedule
-- This defines WHEN a doctor is generally available
-- e.g. "Every Monday 9am-5pm, Every Tuesday 10am-4pm"
CREATE TABLE doctor_availability (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id   UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL,  -- 0=Sunday, 1=Monday, ..., 6=Saturday
    start_time  TIME NOT NULL, -- e.g. 09:00:00
    end_time    TIME NOT NULL, -- e.g. 17:00:00
    slot_duration_minutes INT NOT NULL DEFAULT 30,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Doctor can only have one schedule per day
    UNIQUE(doctor_id, day_of_week)
);

-- Specific bookable appointment slots
-- Generated from doctor_availability
-- e.g. "Dr. Amara, Monday June 2nd, 10:00am-10:30am"
CREATE TABLE appointment_slots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id   UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    start_time  TIME NOT NULL,
    end_time    TIME NOT NULL,
    is_booked   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- No duplicate slots for same doctor at same time
    UNIQUE(doctor_id, date, start_time)
);

-- Confirmed appointments
CREATE TABLE appointments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id       UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    slot_id         UUID NOT NULL REFERENCES appointment_slots(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending → confirmed → completed → cancelled
    type            VARCHAR(20) NOT NULL DEFAULT 'consultation',
    -- consultation, follow_up, emergency
    notes           TEXT, -- patient's reason for visit
    doctor_notes    TEXT, -- doctor fills this after consultation
    amount          DECIMAL(10,2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One slot can only have one appointment
    UNIQUE(slot_id)
);

-- Payment records
CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id      UUID NOT NULL REFERENCES appointments(id),
    patient_id          UUID NOT NULL REFERENCES patients(id),
    amount              DECIMAL(10,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'NGN',
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending → successful → failed → refunded
    paystack_reference  VARCHAR(100) UNIQUE,
    paystack_access_code VARCHAR(100),
    paid_at             TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_appointment_slots_doctor_date ON appointment_slots(doctor_id, date);
CREATE INDEX idx_appointment_slots_is_booked   ON appointment_slots(is_booked);
CREATE INDEX idx_appointments_patient          ON appointments(patient_id);
CREATE INDEX idx_appointments_doctor           ON appointments(doctor_id);
CREATE INDEX idx_appointments_status           ON appointments(status);
CREATE INDEX idx_payments_appointment          ON payments(appointment_id);
CREATE INDEX idx_payments_reference            ON payments(paystack_reference);
