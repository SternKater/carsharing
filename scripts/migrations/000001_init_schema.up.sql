-- Trigger function for update
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_value VARCHAR(255) UNIQUE NOT NULL, -- UUID
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    user_identify_string VARCHAR(255) NOT NULL
);

CREATE TABLE billing_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT, 
    operation_type VARCHAR(20) NOT NULL, -- 'deposit'/'withdraw'
    amount_penny INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE user_balances (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    amount_penny INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- DB-level check: no negative amount
    CONSTRAINT check_positive_balance CHECK (amount_penny >= 0) 
);

-- Trigger
CREATE TRIGGER update_user_balances_changetimestamp
BEFORE UPDATE ON user_balances
FOR EACH ROW
EXECUTE PROCEDURE update_modified_column();

CREATE TABLE cars (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    car_name VARCHAR(50) NOT NULL,
    car_number VARCHAR(20) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'free' NOT NULL, -- 'free'/'rented'/'maintenance'
    price_per_minute INT NOT NULL
);

CREATE TABLE rentals (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    car_id BIGINT NOT NULL REFERENCES cars(id) ON DELETE RESTRICT,
    
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE, -- NULL until the end
    
    price_per_minute INT NOT NULL, -- on pennies, constant (avoid fro change price due driving)
    total_price_penny INT DEFAULT 0 NOT NULL, 
    
    status VARCHAR(20) DEFAULT 'active' NOT NULL, -- 'active'/'finished'/'canceled'
    
    CONSTRAINT check_prices_positive CHECK (price_per_minute >= 0 AND total_price_penny >= 0),
    CONSTRAINT check_dates_order CHECK (finished_at IS NULL OR finished_at >= started_at)
);

CREATE UNIQUE INDEX idx_rentals_user_active ON rentals(user_id) WHERE status = 'active';
CREATE INDEX idx_billing_history_user_time ON billing_history(user_id, created_at DESC);
CREATE UNIQUE INDEX idx_rentals_car_active ON rentals(car_id) WHERE status = 'active';
