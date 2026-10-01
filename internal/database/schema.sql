CREATE TABLE types(
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL UNIQUE
);

CREATE TABLE classes(
    id BIGSERIAL PRIMARY KEY,
    class TEXT NOT NULL UNIQUE
);

CREATE TABLE resource_records (
    id BIGSERIAL PRIMARY KEY,
    domain TEXT NOT NULL,
    data TEXT NOT NULL,
    type_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    time_to_live BIGINT DEFAULT 0,
    FOREIGN KEY (type_id) REFERENCES types(id),
    FOREIGN KEY (class_id) REFERENCES classes(id),
    UNIQUE(domain, data, type_id, class_id) 
);

CREATE TABLE roles(
    id BIGSERIAL PRIMARY KEY,
    role VARCHAR(20) NOT NULL UNIQUE
);

CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(16) NOT NULL UNIQUE,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    password VARCHAR(72) NOT NULL,
    role_id INTEGER NOT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);
