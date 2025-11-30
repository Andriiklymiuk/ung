-- Add agentic early-exit fields to dig_sessions
ALTER TABLE dig_sessions ADD COLUMN early_exit BOOLEAN DEFAULT FALSE;
ALTER TABLE dig_sessions ADD COLUMN early_exit_reason TEXT;
ALTER TABLE dig_sessions ADD COLUMN viability_check TEXT;
ALTER TABLE dig_sessions ADD COLUMN pivot_focus BOOLEAN DEFAULT FALSE;
ALTER TABLE dig_sessions ADD COLUMN flaw_type TEXT;
