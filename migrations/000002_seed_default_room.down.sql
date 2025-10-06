-- Remove the default room
-- This will cascade delete all room_user_settings for this room

DELETE FROM rooms WHERE id = 'c0000000-0000-0000-0000-000000000001';
