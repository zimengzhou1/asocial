-- Seed the default public chat room
-- Using a fixed UUID for consistency across environments

INSERT INTO rooms (id, name, slug, description, owner_id, is_public, created_at, updated_at)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'General',
    'general',
    'The default public chat room. Everyone is welcome!',
    NULL,  -- No owner (system room)
    true,
    NOW(),
    NOW()
);
