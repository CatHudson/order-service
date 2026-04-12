-- +migrate Up
create table if not exists orders_audit_log (
    id            uuid         primary key,
    order_id      uuid         not null references orders(id),
    action        text         not null,
    payload       jsonb        not null,
    created_at    timestamptz   not null default now()
);

-- +migrate Down
drop table if exists orders_audit_log;
