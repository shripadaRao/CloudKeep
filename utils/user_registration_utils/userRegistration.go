package user_registration_utils

import (
	"CloudKeep/models"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string) {
	hashedSaltedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Panic(err)
		return ""
	}
	return string(hashedSaltedPassword)
}

func GenerateSalt() string {
	u, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return u.String()
}

func CreateVerifiedUserRegistration(db *sql.DB, table string, user models.User) error {
	fmt.Println(user.Salt)
    query := `
        INSERT INTO "` + table + `" (username, password, userEmail, salt, userId)
        VALUES ($1, $2, $3, $4, $5)
    `

    _, err := db.Exec(query, user.Username, user.Password, user.UserEmail, user.Salt, user.UserID)
    if err != nil {
        log.Printf("Error inserting data: %v\n", err)
        return err
    }

    return nil
}
