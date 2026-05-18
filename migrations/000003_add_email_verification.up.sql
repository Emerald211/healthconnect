ALTER TABLE patients ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE doctors  ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT false;
