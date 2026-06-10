DROP INDEX IF EXISTS idx_rentals_car_active;
DROP INDEX IF EXISTS idx_rentals_user_active;
DROP INDEX IF EXISTS idx_billing_history_user_time;

DROP TABLE IF EXISTS rentals;
DROP TABLE IF EXISTS cars;
DROP TABLE IF EXISTS user_balances;
DROP TABLE IF EXISTS billing_history;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS update_modified_column;
