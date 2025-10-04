-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS room_user_settings;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";
