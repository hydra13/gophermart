-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number TEXT NOT NULL,
    sum BIGINT NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'PENDING'
);

-- Хак: ADD CONSTRAINT не поддерживает IF NOT EXISTS
ALTER TABLE withdrawals DROP CONSTRAINT IF EXISTS chk_withdrawal_sum_positive;
ALTER TABLE withdrawals ADD CONSTRAINT chk_withdrawal_sum_positive
    CHECK (sum > 0 AND sum <= 999999999999);

CREATE UNIQUE INDEX IF NOT EXISTS idx_withdrawals_user_order ON withdrawals (user_id, order_number);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_order_number ON withdrawals(order_number);
CREATE INDEX IF NOT EXISTS idx_withdrawals_status ON withdrawals(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
