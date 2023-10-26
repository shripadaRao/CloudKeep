CREATE TABLE IF NOT EXISTS my_table ( 
    id serial PRIMARY KEY,
    name VARCHAR(255),
    age INT
);

CREATE TABLE IF NOT EXISTS user_table (
    userId     VARCHAR  PRIMARY KEY,
    userName   VARCHAR,
    password   VARCHAR,
    salt       VARCHAR,
    userEmail  VARCHAR
);
