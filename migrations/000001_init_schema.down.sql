-- Drop tables in reverse order (respecting foreign key constraints)
-- CASCADE will drop dependent objects (like oauth_accounts if it exists)
DROP TABLE IF EXISTS room_user_settings CASCADE;
DROP TABLE IF EXISTS rooms CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";
