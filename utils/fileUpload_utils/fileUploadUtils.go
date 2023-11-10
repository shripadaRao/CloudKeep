package fileUpload_utils

import (
	"CloudKeep/models"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	_ "github.com/lib/pq"
)

func PreUploadMetadataToVideoTable(uploadInitResponse models.UploadInitializationResponse, db *sql.DB) error{
	STATUS := "PENDING"
	query := `INSERT INTO video (video_id, user_id, status, total_chunks, video_path) 
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(query, uploadInitResponse.VideoID, uploadInitResponse.UserID, STATUS, len(uploadInitResponse.ChunkFileNames), "")
	if err != nil {
		fmt.Printf("error writing to video table: %v", err)
		return err
	}

	return nil
}

func PreUploadMetadataToVideoChunksTable(uploadInitResponse models.UploadInitializationResponse, db *sql.DB) error{
	STATUS := "NOT-UPLOADED"

	query := `INSERT INTO video_chunks (chunk_id, video_id, chunk_number, status, chunk_path, check_sum) 
			VALUES ($1, $2, $3, $4, $5, $6)`

	for index, value := range uploadInitResponse.ChunkFileNames {
		_, err := db.Exec(query, value, uploadInitResponse.VideoID, index, STATUS, "", uploadInitResponse.CheckSumMap[value])
		if err != nil {
			fmt.Printf("error writing to video chunks table: %v", err)
			return err
		}		
	}
	return nil
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

func CalculateCompareSHA256(filepath string, chunkCheckSum string) bool {
	sha256sum, err := calculateSHA256(filepath)
	if err != nil {
		fmt.Printf("Error in calculating SHA256 of chunk file %v: %v\n", filepath, err)
		return false
	}

	return sha256sum == chunkCheckSum
}



// func GetVideoDetails(db *sql.DB, chunkID string) (*models.ChunkVerificationDetails, error) {
// 	var chunkVerificationDetails models.ChunkVerificationDetails

// 	query := `
// 		SELECT vc.chunk_id, vc.video_id, v.user_id, vc.check_sum
// 		FROM video_chunks vc
// 		JOIN video v ON vc.video_id = v.video_id
// 		WHERE vc.chunk_id = ?
// 	`

// 	row := db.QueryRow(query, chunkID)

// 	err := row.Scan(&chunkVerificationDetails.ChunkID, &chunkVerificationDetails.VideoID, &chunkVerificationDetails.UserID, &chunkVerificationDetails.Checksum)

// 	if err == sql.ErrNoRows {
// 		return models.ChunkVerificationDetails{}, fmt.Errorf("no record found for chunk_id: %s", chunkID)
// 	} else if err != nil {
// 		return models.ChunkVerificationDetails{}, err
// 	}

// 	return chunkVerificationDetails, nil
// }

func GetVideoDetails(db *sql.DB, chunkID string) (*models.ChunkVerificationDetails, error) {
	var chunkVerificationDetails models.ChunkVerificationDetails

	query := `
		SELECT vc.chunk_id, vc.video_id, v.user_id, vc.check_sum
		FROM video_chunks vc
		JOIN video v ON vc.video_id = v.video_id
		WHERE vc.chunk_id = $1
	`

	row := db.QueryRow(query, chunkID)

	err := row.Scan(&chunkVerificationDetails.ChunkID, &chunkVerificationDetails.VideoID, &chunkVerificationDetails.UserID, &chunkVerificationDetails.Checksum)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no record found for chunk_id: %s", chunkID)
		}
		return nil, err
	}

	return &chunkVerificationDetails, nil
}