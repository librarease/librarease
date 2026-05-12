-- Create notification trigger function. Live streams are recipient-scoped so
-- each SSE client receives only inbox rows that belong to its authenticated user.
CREATE OR REPLACE FUNCTION notify_new_notification_recipient()
RETURNS TRIGGER AS $$
DECLARE
    event_row notification_events%ROWTYPE;
BEGIN
    SELECT * INTO event_row FROM notification_events WHERE id = NEW.event_id;

    PERFORM pg_notify(
        'new_notification',
        json_build_object(
            'ID', NEW.id,
            'EventID', NEW.event_id,
            'UserID', NEW.user_id,
            'Title', event_row.title,
            'Message', event_row.message,
            'CreatedAt', NEW.created_at,
            'UpdatedAt', NEW.updated_at,
            'ReadAt', NEW.read_at,
            'ReferenceID', event_row.reference_id,
            'ReferenceType', event_row.reference_type
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF to_regclass('public.notifications') IS NOT NULL THEN
        INSERT INTO notification_events (
            id,
            title,
            message,
            reference_id,
            reference_type,
            created_at,
            updated_at,
            deleted_at
        )
        SELECT
            n.id,
            n.title,
            n.message,
            n.reference_id,
            n.reference_type,
            n.created_at,
            n.updated_at,
            n.deleted_at
        FROM notifications n
        WHERE NOT EXISTS (
            SELECT 1 FROM notification_events e WHERE e.id = n.id
        );

        INSERT INTO notification_recipients (
            id,
            event_id,
            user_id,
            read_at,
            sent_at,
            created_at,
            updated_at,
            deleted_at
        )
        SELECT
            n.id,
            n.id,
            n.user_id,
            n.read_at,
            n.sent_at,
            n.created_at,
            n.updated_at,
            n.deleted_at
        FROM notifications n
        WHERE NOT EXISTS (
            SELECT 1 FROM notification_recipients r WHERE r.id = n.id
        );
    END IF;
END;
$$;

DO $$
BEGIN
    IF to_regclass('public.notifications') IS NOT NULL THEN
        DROP TRIGGER IF EXISTS notify_new_notification_trigger ON notifications;
    END IF;
END;
$$;
DROP TRIGGER IF EXISTS notify_new_notification_recipient_trigger ON notification_recipients;
CREATE TRIGGER notify_new_notification_recipient_trigger
    AFTER INSERT ON notification_recipients
    FOR EACH ROW
    EXECUTE FUNCTION notify_new_notification_recipient();
