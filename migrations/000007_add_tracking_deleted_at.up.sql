-- Add deleted_at column to tracking_sessions for soft delete support
ALTER TABLE tracking_sessions ADD COLUMN deleted_at TIMESTAMP;
