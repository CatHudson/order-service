-- +migrate Up
create table if not exists orders (
    id                          uuid primary key,
    account_id                  uuid not null,
    idempotency_key             uuid not null unique,
    instrument_id               uuid not null,
    order_by                    varchar(50) not null check (order_by in ('QUANTITY', 'AMOUNT')),
    side                        varchar(50) not null check (side in ('BUY', 'SELL')),
    amount                      numeric(18, 10),
    quantity                    numeric(18, 10),
    price                       numeric(18, 10),
    status                      varchar(50) not null check (status in ('NEW', 'PENDING', 'SUCCESS', 'FAILED', 'CANCELED')),
    error_message               text,
    created_at                  timestamptz not null default now(),
    updated_at                  timestamptz not null default now()
);

-- +migrate Down
drop table if exists orders;
