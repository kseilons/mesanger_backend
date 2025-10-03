-- Drop groups related tables
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS update_channels_updated_at ON channels;
DROP TABLE IF EXISTS channel_members;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
