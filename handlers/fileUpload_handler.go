package handlers

import (
	"CloudKeep/models"
	"CloudKeep/utils/fileUpload_utils"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SimpleTestUploadAPI(c *gin.Context, db *sql.DB) {
	file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filePath := getDestinationPath(file.Filename, true)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			fmt.Println("failed to upload file. filename: ", file.Filename)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

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
	fmt.Println("file uploading initialized for, filename: ",uploadInitResponse.VideoID )
	c.JSON(http.StatusAccepted, gin.H{"message": "Upload pipeline is initialized", "error": nil})

}

func getDestinationPath(filename string, isMP4 bool) string {
	tempDir := os.TempDir()
	destinationPath := filepath.Join(tempDir, filename)
	if !isMP4{
	return destinationPath }
	return destinationPath + ".mp4"
}

func UploadChunk(c *gin.Context, db *sql.DB) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read multipart form", "error": err.Error()})
		return
	}

	var chunkVerificationDetails models.ChunkVerificationDetails
	chunkID := form.Value["chunkID"][0]
	fmt.Println("uploading chunkID: ",chunkID);
	
	chunkVerificationDetails, err = fileUpload_utils.GetChunkDetails(db, chunkID)
	if err != nil {
		fmt.Println("Failed to fetch chunk details. chunkID: ", chunkID)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch chunk details", "error": err.Error()})
		return
	}

    //to-do check whether userid and vid match through jwt and chunkVerificationDetails



	// Assuming "chunkFile" is the name of the file input field in your form
	chunkFile, exists := form.File["chunkFile"]
	if !exists || len(chunkFile) == 0 {
		fmt.Println("failed to recieve/read chunk file from form data. ChunkID: ", chunkID)
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to recieve/read chunk file from form data"})
		return
	}

	err = c.SaveUploadedFile(chunkFile[0], getDestinationPath(chunkFile[0].Filename, false))
	if err != nil {
		fmt.Println("Failed to store chunk file. chunkID", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store chunk file", "error": err.Error()})
		return
	}

	// set chunk_path in db
	err = fileUpload_utils.UpdateFieldInVideoChunksTable(db, chunkID, "chunk_path", getDestinationPath(chunkFile[0].Filename, false))
	if err != nil {
		fmt.Println(fmt.Sprintf("error in updating status in chunk %v", chunkID), err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("error in updating status in chunk %v", chunkID), "error": err.Error()})
		return
	}

	//checksum of the file received
	if !fileUpload_utils.CalculateCompareSHA256(getDestinationPath(chunkFile[0].Filename, false), chunkVerificationDetails.Checksum) {
		fmt.Println("chunk checksum failed. chunkID: ", chunkID)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Checksum failed"})
		return	}

	// set status codes of chunks as IN-SERVER
	err = fileUpload_utils.UpdateFieldInVideoChunksTable(db, chunkID, "status", "IN-SERVER")
	if err != nil {
		fmt.Println(fmt.Sprintf("error in updating status in chunk %v", chunkID), err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("error in updating status in chunk %v", chunkID), "error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": fmt.Sprintf("Chunkfile %v uploaded successfully", chunkFile[0].Filename), "error": nil})
}

func MergeChunks(c *gin.Context, db *sql.DB) {

	type MergeChunksModel struct {
		VideoId string  `json:"video_id"`
	}
	var  videoData MergeChunksModel
	if err := c.ShouldBindJSON(&videoData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message":"error in recieving response body"})
		return
	}

	// construct chunkFilenamesByOrder arr
	chunkFileNamesInOrder, _ := fileUpload_utils.GetVideoSequencedChunks(db, videoData.VideoId)
	var chunkFilePathsInOrder []string
	for _, fileName := range chunkFileNamesInOrder {
		chunkFilePath := fmt.Sprintf("/tmp/%s", fileName)
		chunkFilePathsInOrder = append(chunkFilePathsInOrder, chunkFilePath)
	}	

	// merge all together
	destinationFilePath := path.Join("/tmp", videoData.VideoId+".mp4")
	fileUpload_utils.MergeChunks(chunkFilePathsInOrder, destinationFilePath)

	// validate checksum for complete video
	var originalFileDetails models.Video
	originalFileDetails, _ = fileUpload_utils.GetFileDetailsFromId(db, videoData.VideoId)
	
	if !fileUpload_utils.CalculateCompareSHA256(destinationFilePath, originalFileDetails.CheckSum) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Checksum failed"})
		return	}	

	// delete chunks
	err := fileUpload_utils.DeleteFiles(chunkFilePathsInOrder...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in deleting chunk files", "error" : err.Error()})
		return
	}

	// set status 
	err = fileUpload_utils.UpdateFieldInVideoTable(db, videoData.VideoId, "status", "COMPLETE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in updating status to complete in video table", "error" : err.Error()})
		return 
	}
	err = fileUpload_utils.UpdateFieldInVideoTable(db, videoData.VideoId, "video_path", destinationFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in updating status to complete in video table", "error" : err.Error()})
		return
	}		

	c.JSON(http.StatusAccepted, gin.H{"message": fmt.Sprintf("merged the chunks and constructed video successfully. %v",destinationFilePath), "error": nil})
}

// kick off this as a background job
func StoreMergedFileS3(c *gin.Context, db *sql.DB) {
	type VideoIDStruct struct {
		VideoID string  `json:"video_id"`
	}
	var  requestObj VideoIDStruct
	if err := c.ShouldBindJSON(&requestObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message":"error in recieving response body"})
		return
	}

	fmt.Println("videoid: ", requestObj.VideoID)

	// check if video belongs to user - postgres query
	var UserID string
	query := `SELECT user_id FROM video WHERE video_id = $1 `
	row := db.QueryRow(query, requestObj.VideoID)
	err := row.Scan(&UserID)
	if err != nil {
		fmt.Println("Error in fetching userID from videoID", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message":"Error in fetching userID from videoID"})
		return
	}
	claims, _ := c.Get("claims")
	jwtUserID := claims.(*models.JWTClaims).UserId

	if UserID != jwtUserID {
		fmt.Println("Mismatch in userID, request failed")
		c.JSON(http.StatusBadRequest, gin.H{"message":"Mismatch in userID, request failed"})
		return
	}

	// push to s3
	s3URI, err := fileUpload_utils.UploadFileToS3(getDestinationPath(requestObj.VideoID, true))
	if err != nil {
		fmt.Println("Error in Uploading video file to s3", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message":"Error in Uploading video file to s3", "error": err.Error()})
		return
	}
	
	// update filepath in postgres
	err = fileUpload_utils.UpdateFieldInVideoTable(db, requestObj.VideoID, "video_path", s3URI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in updating status to complete in video table", "error" : err.Error()})
		return 
	}

	// clean up disk - delete file from /tmp
	err = os.Remove(getDestinationPath(requestObj.VideoID, true))
	if err != nil {
		fmt.Println("Error deleting file: ", err)
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "video file successfully pushed to s3 and removed from /tmp", "error": nil})

}