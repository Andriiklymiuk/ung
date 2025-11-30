-- Remove agentic early-exit fields from dig_sessions
ALTER TABLE dig_sessions DROP COLUMN IF EXISTS early_exit;
ALTER TABLE dig_sessions DROP COLUMN IF EXISTS early_exit_reason;
ALTER TABLE dig_sessions DROP COLUMN IF EXISTS viability_check;
ALTER TABLE dig_sessions DROP COLUMN IF EXISTS pivot_focus;
ALTER TABLE dig_sessions DROP COLUMN IF EXISTS flaw_type;
