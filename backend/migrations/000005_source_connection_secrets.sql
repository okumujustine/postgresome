ALTER TABLE source_connection_profiles
    ADD COLUMN IF NOT EXISTS connection_uri_encrypted TEXT NOT NULL DEFAULT '';

-- Legacy plaintext rows may still exist in development databases.
-- New writes should use connection_uri_encrypted and leave connection_uri empty.
