-- This is the SQL script that will be used to initialize the database schema.
-- We will evaluate you based on how well you design your database.
-- 1. How you design the tables.
-- 2. How you choose the data types and keys.
-- 3. How you name the fields.
-- In this assignment we will use PostgreSQL as the database.


CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- all number that more than  32,767 using INTEGER (4 bytes) -2,147,483,648 to +2,147,483,647
-- we cannot use the smaller type : SMALLINT (2 bytes, -32,768 to 32,767) because it could not stored if 50,000 be stored.
-- totalDistance also still possible to Using INTEGER 50,000 x 50,000 x 10 x (30 x 2) = 1,500,000 which still more than enough to use INTEGER.
CREATE TABLE estates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    width INTEGER NOT NULL CHECK (width >= 1 AND width <= 50000),
    length INTEGER NOT NULL CHECK (length >= 1 AND length <= 50000),
    tree_count INTEGER NOT NULL CHECK (tree_count >= 0 AND tree_count <= 1000),
    tree_max_height SMALLINT NOT NULL CHECK (tree_max_height >= 0 AND tree_max_height <= 30),
    tree_min_height SMALLINT NOT NULL CHECK (tree_min_height >= 0 AND tree_min_height <= 30),
    tree_median_height SMALLINT NOT NULL CHECK (tree_median_height >= 0 AND tree_median_height <= 30),
    total_distance INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE plots (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     x INTEGER NOT NULL CHECK (x >= 1 AND x <= 50000),
     y INTEGER NOT NULL CHECK (y >= 1 AND y <= 50000),
     estate_id UUID NOT NULL,
     order_number INTEGER NOT NULL,
     tree_height SMALLINT NOT NULL CHECK (tree_height >= 1 AND tree_height <= 30),
     distance INTEGER NOT NULL,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     FOREIGN KEY (estate_id) REFERENCES estates(id)
);

CREATE INDEX idx_estate_id ON plots (estate_id);
CREATE INDEX idx_estate_id_order_number ON plots (estate_id, order_number);
CREATE INDEX idx_x_y ON plots (x, y);