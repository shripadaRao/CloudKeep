package database

import (
	"database/sql"
	"fmt"
)

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func makeConnection(config DatabaseConfig) (*sql.DB, error) {
    connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}
