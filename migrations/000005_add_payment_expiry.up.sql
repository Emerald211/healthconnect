ALTER TABLE payments ADD COLUMN expires_at TIMESTAMPTZ;

-- Set expiry to 30 minutes from now for all existing pending payments
UPDATE payments SET expires_at = created_at + INTERVAL '30 minutes' WHERE status = 'pending';
