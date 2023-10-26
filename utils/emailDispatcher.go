package utils

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func ParseEmailTemplate(filename string) (emailSubjectTemplate string, emailBodyTemplate string, error error){
    var emailTemplate map[string]interface{}

    file, err := os.Open(filename)
    if err != nil {
		log.Fatalln("error while opening the config file in parseEmailTemplate")
        return "", "", err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    decodingError := decoder.Decode(&emailTemplate)
	if decodingError != nil {
		log.Fatalln("error in decoding json data in parseEmailTemplate")
        return "", "", decodingError
    }

	return emailTemplate["subject"].(string), emailTemplate["body"].(string), decodingError

}

func SendEmail(receiptEmail string, emailSubject string, emailBody string) (isSuccess bool, err error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file while sending email: %v", err)
	}

	ownerEmail := os.Getenv("OWNER_EMAIL")
	ownerEmailPassword := os.Getenv("OWNER_EMAIL_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", ownerEmail)
	m.SetHeader("To", receiptEmail)
	m.SetHeader("Subject", emailSubject)
	m.SetBody("text/plain", emailBody)

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		ownerEmail,
		ownerEmailPassword,
	)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Email Dispatch failed: ", err)
		isSuccess := false
		return isSuccess, err
	} else {
		fmt.Println("Email sent successfully!")
		isSuccess := true
		return isSuccess, nil
	}
}

func GenerateOTP() (string) {
	min := big.NewInt(100000)
	max := big.NewInt(999999)

	randomNum, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		fmt.Println("Error generating OTP: ", err)
		return ""
	}

	otp := randomNum.Add(randomNum, min)

	return fmt.Sprintf("%06d", otp)
}

func GenerateEmailBody(emailBodyTemplate string, otp string) string{
	return emailBodyTemplate + "\n" + otp
}