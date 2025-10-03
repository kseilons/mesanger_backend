-- Drop notifications related tables and functions
DROP FUNCTION IF EXISTS get_unread_notification_count(UUID);
DROP FUNCTION IF EXISTS mark_notifications_as_read(UUID, UUID[]);
DROP TABLE IF EXISTS notifications;
