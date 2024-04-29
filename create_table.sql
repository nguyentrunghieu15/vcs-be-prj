CREATE TYPE ROLE AS ENUM ('admin', 'user');
CREATE TYPE SERVER_STATUS AS ENUM ('on', 'off');
CREATE TYPE FILE_STATUS AS ENUM ('exported', 'exporting','fail');
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    created_by INT,
    updated_at TIMESTAMP,
    updated_by INT,
    deleted_at TIMESTAMP,
    deleted_by INT,
    email VARCHAR(255) NOT NULL,
    UNIQUE (email,deleted_at),
    full_name VARCHAR(1000),
    password VARCHAR(2000),
    phone VARCHAR(20),
    avatar VARCHAR(2000),
    is_supper_admin BOOLEAN,
    roles ROLE
);

CREATE TABLE servers (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY ,
    created_at TIMESTAMP,
    created_by INT,
    updated_at TIMESTAMP,
    updated_by INT,
    deleted_at TIMESTAMP,	
    deleted_by INT,
    "name" VARCHAR(2000) NOT NULL,
    UNIQUE ("name",deleted_at),
    status SERVER_STATUS,
    ipv4 VARCHAR(16),
    foreign key (created_by) references users(id)
);

CREATE TABLE files (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY ,
    created_at TIMESTAMP,
    created_by INT,
    updated_at TIMESTAMP,
    updated_by INT,
    deleted_at TIMESTAMP,
    deleted_by INT,
    file_name VARCHAR(1000),
    file_path VARCHAR(1000),
    file_size INT,
    status FILE_STATUS,
    owner INT,
    foreign key (owner) references users(id)
);