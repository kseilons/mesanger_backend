-- Drop messages related tables and functions
DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
DROP FUNCTION IF EXISTS get_message_thread(UUID);
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS message_attachments;
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS messages;
