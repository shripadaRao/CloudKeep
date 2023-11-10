package handlers

import (
	"CloudKeep/models"
	"CloudKeep/utils/fileUpload_utils"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func InitializeUploadProcess(c *gin.Context, db *sql.DB) {
	var uploadInitResponse models.UploadInitializationResponse
	if err := c.ShouldBindJSON(&uploadInitResponse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message":"error in recieving response body"})
		return
	}

	err := fileUpload_utils.PreUploadMetadataToVideoTable(uploadInitResponse, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error writing to video table", "error": err.Error()})
		return
	}

	err = fileUpload_utils.PreUploadMetadataToVideoChunksTable(uploadInitResponse, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error writing to video chunks table", "error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Upload pipeline is initialized", "error": nil})

}

func getDestinationPath(filename string) string {
	tempDir := os.TempDir()
	destinationPath := filepath.Join(tempDir, filename)
	fmt.Println("chunk saved at: ",destinationPath)
	return destinationPath
}

func UploadChunk(c *gin.Context, db *sql.DB) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read multipart form", "error": err.Error()})
		return
	}

	//to-do check whether userid and vid match
	var chunkVerificationDetails models.ChunkVerificationDetails
	chunkID := form.Value["chunkID"][0]
	fmt.Println(chunkID);
	
	details, err := fileUpload_utils.GetVideoDetails(db, chunkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch chunk details", "error": err.Error()})
		return
	}
	
	chunkVerificationDetails = *details
	fmt.Println("chunk id: ",chunkVerificationDetails.ChunkID)


	// Assuming "chunkFile" is the name of the file input field in your form
	chunkFile, exists := form.File["chunkFile"]
	if !exists || len(chunkFile) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to recieve/read chunk file from form data"})
		return
	}

	err = c.SaveUploadedFile(chunkFile[0], getDestinationPath(chunkFile[0].Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store chunk file", "error": err.Error()})
		return
	}

	//check md5sum of the file received
	if !fileUpload_utils.CalculateCompareSHA256(getDestinationPath(chunkFile[0].Filename), chunkVerificationDetails.Checksum) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Checksum failed"})
		return	}


	c.JSON(http.StatusAccepted, gin.H{"message": fmt.Sprintf("Chunkfile %v uploaded successfully", chunkFile[0].Filename), "error": nil})
}

