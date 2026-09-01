-- +goose Up
-- table users
CREATE TABLE IF NOT EXISTS users(
    inn TEXT NOT NULL,
    company_id TEXT NOT NULL,
    user_login TEXT NOT NULL,
    password_kit TEXT NOT NULL,
    password_fns TEXT NOT NULL,
    device_id TEXT NOT NULL,
    time_zone TEXT NOT NULL,
    time_offset INTEGER NOT NULL,
    refresh_token_fns TEXT NOT NULL,
    token_fns TEXT NOT NULL,
    auto_mode BOOLEAN NOT NULL,
    is_paid BOOLEAN NOT NULL,
    PRIMARY KEY(inn, company_id, user_login)
    );

CREATE TABLE IF NOT EXISTS automode_errors(
    inn TEXT NOT NULL,
    company_id TEXT NOT NULL,
    user_login TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    message TEXT NOT NULL,
    is_business_err BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS status_errors(
    inn TEXT NOT NULL,
    company_id TEXT NOT NULL,
    user_login TEXT NOT NULL,
    is_solved BOOLEAN NOT NULL,
    PRIMARY KEY(inn, company_id, user_login)
    );
-- +goose Down