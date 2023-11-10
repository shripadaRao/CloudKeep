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

CREATE TABLE IF NOT EXISTS video (
    video_id VARCHAR PRIMARY KEY,
    user_id VARCHAR,
    status VARCHAR CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED')),
    total_chunks INT, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    video_path VARCHAR   
);

CREATE TABLE IF NOT EXISTS video_chunks (
    chunk_id VARCHAR PRIMARY KEY,
    video_id  VARCHAR,
    chunk_number    INT,
    status VARCHAR CHECK (status IN ('NOT-UPLOADED', 'IN-SERVER', 'IN-S3')),
    chunk_path VARCHAR, 
    check_sum VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);

