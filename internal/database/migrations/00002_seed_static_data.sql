-- +goose Up
-- +goose StatementBegin
INSERT INTO types (type) VALUES
    ('A'),
    ('AAAA'),
    ('CNAME'),
    ('MX'),
    ('NS'),
    ('SOA'),
    ('TXT'),
    ('SRV')
ON CONFLICT (type) DO NOTHING;

INSERT INTO classes (class) VALUES
    ('IN'),   -- Internet
    ('CH'),   -- Chaos
    ('HS')    -- Hesiod
ON CONFLICT (class) DO NOTHING;

INSERT INTO roles (role) VALUES
    ('admin'),
    ('user')
ON CONFLICT (role) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM types WHERE type IN ('A','AAAA','CNAME','MX','NS','SOA','TXT','SRV');
DELETE FROM classes WHERE class IN ('IN','CH','HS');
DELETE FROM roles WHERE role IN ('admin','user');
-- +goose StatementEnd
