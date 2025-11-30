-- Drop Dig feature tables
DROP INDEX IF EXISTS idx_dig_alternatives_session;
DROP INDEX IF EXISTS idx_dig_revenue_session;
DROP INDEX IF EXISTS idx_dig_marketing_session;
DROP INDEX IF EXISTS idx_dig_execution_session;
DROP INDEX IF EXISTS idx_dig_analyses_perspective;
DROP INDEX IF EXISTS idx_dig_analyses_session;
DROP INDEX IF EXISTS idx_dig_sessions_created;
DROP INDEX IF EXISTS idx_dig_sessions_status;

DROP TABLE IF EXISTS dig_alternatives;
DROP TABLE IF EXISTS dig_revenue_projections;
DROP TABLE IF EXISTS dig_marketing;
DROP TABLE IF EXISTS dig_execution_plans;
DROP TABLE IF EXISTS dig_analyses;
DROP TABLE IF EXISTS dig_sessions;
