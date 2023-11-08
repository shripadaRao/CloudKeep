package handlers

import (
	"CloudKeep/models"
	"CloudKeep/utils/email_utils"
	"CloudKeep/utils/redis_utils"
	"CloudKeep/utils/user_registration_utils"

	// "CloudKeep/utils"
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


func SendRegistrationEmail(c *gin.Context, ctx context.Context, redisClient *redis.Client) {
    var requestData struct {
        UserEmail string `json:"userEmail"`
        UserId    string `json:"userId"`
    }

    if err := c.ShouldBindJSON(&requestData); err != nil {
        log.Printf("Error in receiving user data: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "error"})
        return
    }

	templateEmailSubject, templateEmailBody, err := email_utils.ParseEmailTemplate("config/registrationEmailTemplate.json")
	if err != nil {
		log.Printf("Error in loading email template: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "error"})
		return 
	}

	otp := email_utils.GenerateOTP()
	emailBody := email_utils.GenerateEmailBody(templateEmailBody, otp)

	isEmailDispatchSuccessful, err := email_utils.SendEmail(requestData.UserEmail, templateEmailSubject, emailBody)
	if err != nil {
		log.Printf("Email dispatch failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "error"})
		return 
	}

	if !isEmailDispatchSuccessful {
		err := fmt.Errorf("Email dispatch failed")
		log.Printf("Email dispatch failed: %v", err)
		return 
	}

    var emailRegistrationRequest models.EmailRegistrationRequest
    emailRegistrationRequest.IsVerified = false
    emailRegistrationRequest.OTP = otp

    err = redis_utils.RedisSetData(ctx, redisClient, requestData.UserId, emailRegistrationRequest, OTPExpirationTime)
    if err != nil {
        log.Printf("Error in writing into redis while sending otp: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error(), "status": "error"})
    }
    c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully", "status": "success"})
	return 
}

func VerifyRegistrationOTP(c *gin.Context, ctx context.Context, redisClient *redis.Client) {
    var userInput models.VerifyRegistrationOTPModel
    if err := c.BindJSON(&userInput); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error:": fmt.Sprintf("error in parsing user input while verifying otp %s: ",err.Error()), "status": "error"})
        return
    }
    // fetch otp from redis key
    userRegistrationData, err := redis_utils.RedisGetData(ctx, redisClient, userInput.UserId)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Key", "message": "Key has been expired or does not exist."})
        return
    }   

    var userData models.VerifyRegistrationOTPModel
    if err := json.Unmarshal([]byte(userRegistrationData.(string)), &userData); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error", "message": "Failed to retrieve user data from Redis."})
        return
    }

    var userEmailVerificationData models.EmailRegistrationRequest
    userEmailVerificationData.IsVerified = true

    if (userInput.OTP == userData.OTP){
            // otp check success
            err := redis_utils.RedisSetData(ctx, redisClient, userInput.UserId, userEmailVerificationData, OTPExpirationTime)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error":"error in setting redis client", "status": "error"})
            }

            c.JSON(http.StatusOK, gin.H{"message": "OTP verification successful", "status": "success"})
            } else {
        // otp check failed
        //return 401
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP", "message": "The provided OTP is incorrect."})
            }
} 

func CreateUser(c *gin.Context, ctx context.Context, redisClient *redis.Client, db *sql.DB) {
    var userInput models.RegisterUser
    if err := c.BindJSON(&userInput); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error:": fmt.Sprintf("error in parsing user input while verifying otp %s: ",err.Error()), "status": "error"})
        return
    }
    //check redis for verification
    userPreRegistrationData, err := redis_utils.RedisGetData(ctx, redisClient, userInput.UserID)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Key", "message": "Key has been expired or does not exist."})
        return
    }   

    var userData models.EmailRegistrationRequest
    if err := json.Unmarshal([]byte(userPreRegistrationData.(string)), &userData); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error", "message": "Failed to retrieve user data from Redis."})
        return
    }

    if userData.IsVerified == false {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to authenticate", "status": "failed"})
        return
    }

    //password processing
    userInputPassword := userInput.Password
    salt := user_registration_utils.GenerateSalt()
    saltedPassword := userInputPassword + salt
    hashedPassword := user_registration_utils.HashPassword(saltedPassword)

    //database entry
    var user models.User
    user.Username = userInput.Username
    user.UserID = userInput.UserID
    user.UserEmail =  userInput.UserEmail
    user.Password = hashedPassword
    user.Salt = salt    

    err = user_registration_utils.CreateVerifiedUserRegistration(db, "user_table", user)
    if err != nil {
        log.Printf("failed to write to database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while writing to db: %v", err), "status": "failed"})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"message":"successful in creating new user", "status": "success"})
    return 
}