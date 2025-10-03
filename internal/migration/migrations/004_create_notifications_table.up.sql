-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('new_message', 'new_reaction', 'group_invite', 'group_update', 'channel_update', 'mention', 'system')),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    data JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;

-- Create function to mark notifications as read
CREATE OR REPLACE FUNCTION mark_notifications_as_read(user_uuid UUID, notification_ids UUID[] DEFAULT NULL)
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    IF notification_ids IS NULL THEN
        -- Mark all notifications for user as read
        UPDATE notifications 
        SET is_read = TRUE, read_at = NOW()
        WHERE user_id = user_uuid AND is_read = FALSE;
        
        GET DIAGNOSTICS updated_count = ROW_COUNT;
    ELSE
        -- Mark specific notifications as read
        UPDATE notifications 
        SET is_read = TRUE, read_at = NOW()
        WHERE user_id = user_uuid AND id = ANY(notification_ids) AND is_read = FALSE;
        
        GET DIAGNOSTICS updated_count = ROW_COUNT;
    END IF;
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to get unread notification count
CREATE OR REPLACE FUNCTION get_unread_notification_count(user_uuid UUID)
RETURNS INTEGER AS $$
DECLARE
    unread_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO unread_count
    FROM notifications
    WHERE user_id = user_uuid AND is_read = FALSE;
    
    RETURN unread_count;
END;
$$ LANGUAGE plpgsql;
