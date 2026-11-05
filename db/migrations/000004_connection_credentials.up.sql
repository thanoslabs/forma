ALTER TABLE connections ADD COLUMN IF NOT EXISTS endpoint_url TEXT;
ALTER TABLE connections ADD COLUMN IF NOT EXISTS credential_ciphertext BYTEA;
