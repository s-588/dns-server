CREATE TABLE types(
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL UNIQUE
);

CREATE TABLE classes(
    id SERIAL PRIMARY KEY,
    class TEXT NOT NULL UNIQUE
);

CREATE TABLE resource_records (
    id SERIAL PRIMARY KEY,
    domain TEXT NOT NULL,
    data TEXT NOT NULL,
    type_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    time_to_live INTEGER DEFAULT 0,
    FOREIGN KEY (type_id) REFERENCES types(id),
    FOREIGN KEY (class_id) REFERENCES classes(id),
    UNIQUE(domain, data, type_id, class_id) 
);

CREATE TABLE roles(
    id SERIAL PRIMARY KEY,
    role VARCHAR(20) NOT NULL UNIQUE
);

CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    login VARCHAR(16) NOT NULL UNIQUE,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    password VARCHAR(72) NOT NULL,
    role_id INTEGER NOT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);
