package handlers

import (
	"CloudKeep/models"
	"CloudKeep/utils"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const OTPExpirationTime = 10 * time.Minute


func SendRegistrationEmail(emailAddress string) (string, error) {
	templateEmailSubject, templateEmailBody, err := utils.ParseEmailTemplate("config/registrationEmailTemplate.json")
	if err != nil {
		log.Printf("Error loading email template: %v", err)
		return "", err
	}
	otp := utils.GenerateOTP()
	emailBody := utils.GenerateEmailBody(templateEmailBody, otp)

	isEmailDispatchSuccessful, err := utils.SendEmail(emailAddress, templateEmailSubject, emailBody)
	if err != nil {
		log.Printf("Email dispatch failed: %v", err)
		return "", err
	}

	if !isEmailDispatchSuccessful {
		err := fmt.Errorf("Email dispatch failed")
		log.Printf("Email dispatch failed: %v", err)
		return "", err
	}

	return otp, nil
}


func RegisterNewAccount(c *gin.Context, ctx context.Context, redisClient *redis.Client) {
    // Collect UnVerifiedUser details from the request
    var unVerifiedUser models.UnVerifiedUser
    if err := c.BindJSON(&unVerifiedUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error in parsing user input while registering": err.Error()})
        return
    }

    // Send OTP and get the OTP value
    OTP, err := SendRegistrationEmail(unVerifiedUser.UserEmail)
    if err != nil {
        log.Printf("Email dispatch failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
        return
    }

    // Create a RegisterUser object and populate it
    registerUser := models.RegisterUser{
        Username:  unVerifiedUser.Username,
        Password:  unVerifiedUser.Password,
        UserEmail: unVerifiedUser.UserEmail,
        OTP:       OTP,
    }

	//create userid
	//todo improve this later
	userId := registerUser.Username


    // Store the RegisterUser object in Redis
    err = utils.RedisSetData(ctx, redisClient, userId, registerUser, OTPExpirationTime)
    if err != nil {
        log.Printf("Error storing user data in Redis: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store user data"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Email with OTP has been successfully sent"})
}

type VerifyRegistrationOTPModel struct {
    UserEmail string
    UserId  string
    OTP    string
}
func VerifyRegistrationOTP(c *gin.Context, ctx context.Context, redisClient *redis.Client, db *sql.DB) {
 var userInput VerifyRegistrationOTPModel
 if err := c.BindJSON(&userInput); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error in parsing user input while registering": err.Error()})
    return
}
// fetch otp from redis key
 userRegistrationData, err := utils.RedisGetData(ctx, redisClient, userInput.UserId)
 if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Key", "message": "Key has been expired or does not exist."})
    return
}   

var userData models.RegisterUser
if err := json.Unmarshal([]byte(userRegistrationData.(string)), &userData); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error", "message": "Failed to retrieve user data from Redis."})
    return
}

 if (userInput.OTP == userData.OTP){
        // otp check success
        // permanent entry in postgres
        // remove entry from redis
        c.JSON(http.StatusOK, gin.H{"message": "OTP verification successful"})
        } else {
    // otp check failed
    //return 401
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP", "message": "The provided OTP is incorrect."})
        }

} 

//dummy API
    func SendRegistrationEmailAPI(c *gin.Context) {
	_, err := SendRegistrationEmail("srao06558@gmail.com")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send registration email"})
        return
    }

    // Respond with a success message
    c.JSON(http.StatusOK, gin.H{"message": "Registration email sent successfully"})}
