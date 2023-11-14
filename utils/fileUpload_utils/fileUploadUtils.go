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
	query := `INSERT INTO video (video_id, user_id, status, total_chunks, video_path, check_sum) 
			VALUES ($1, $2, $3, $4, $5, $6)`

	fmt.Println("video checksum", uploadInitResponse.FileCheckSum)
	_, err := db.Exec(query, uploadInitResponse.VideoID, uploadInitResponse.UserID, STATUS, len(uploadInitResponse.ChunkFileNames), "", uploadInitResponse.FileCheckSum)
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

func CalculateSHA256(filePath string) (string, error) {
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
	sha256sum, err := CalculateSHA256(filepath)
	if err != nil {
		fmt.Printf("Error in calculating SHA256 of chunk file %v: %v\n", filepath, err)
		return false
	}

	return sha256sum == chunkCheckSum
}


func GetChunkDetails(db *sql.DB, chunkID string) (models.ChunkVerificationDetails, error) {
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
			return chunkVerificationDetails, fmt.Errorf("no record found for chunk_id: %s", chunkID)
		}
		return chunkVerificationDetails, err
	}

	return chunkVerificationDetails, nil
}

func GetVideoSequencedChunks(db *sql.DB, video_id string) ([]string, error) {
	type VideoChunkData struct {
		ChunkID      string
		ChunkNumber  int
	}
	var videoChunkData VideoChunkData
	var chunkFileNamesInOrder []string

	query := `SELECT chunk_id FROM video_chunks WHERE video_id = $1 ORDER BY chunk_number ASC;`

	rows, err := db.Query(query, video_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&videoChunkData.ChunkID)
		if err != nil {
			return nil, err
		}
		chunkFileNamesInOrder = append(chunkFileNamesInOrder, videoChunkData.ChunkID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chunkFileNamesInOrder, nil
}


func MergeChunks(files []string, outputFile string) {
	// Open output file
	output, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	// Merge chunks directly into the output file
	for _, file := range files {
		chunk, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(output, chunk); err != nil {
			chunk.Close()
			panic(err)
		}
		chunk.Close()
	}

	fmt.Println("Merging complete.")
}

func GetFileDetailsFromId(db *sql.DB, video_id string) (models.Video, error) {
	var video models.Video

	query := `
		SELECT * FROM video
		WHERE video_id = $1
	`

	err := db.QueryRow(query, video_id).Scan(
		&video.VideoID,
		&video.UserID,
		&video.Status,
		&video.TotalChunks,
		&video.CreatedAt,
		&video.VideoPath,
		&video.CheckSum,
	)

	if err != nil {
		return video, err
	}

	return video, nil
}


func DeleteFiles(filePaths ...string) error {
	for _, filePath := range filePaths {
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("error deleting file %s: %w", filePath, err)
		}
	}
	return nil
}

func UpdateFieldInVideoTable(db *sql.DB, videoID, fieldName string, newValue interface{}) error {
	query := fmt.Sprintf("UPDATE video SET %s = $1 WHERE video_id = $2", fieldName)

	_, err := db.Exec(query, newValue, videoID)
	if err != nil {
		return fmt.Errorf("error updating field %s: %w", fieldName, err)
	}
	return nil
}

func UpdateFieldInVideoChunksTable(db *sql.DB, chunkID, fieldName string, newValue interface{}) error {
	query := fmt.Sprintf("UPDATE video_chunks SET %s = $1 WHERE video_id = $2", fieldName)

	_, err := db.Exec(query, newValue, chunkID)
	if err != nil {
		return fmt.Errorf("error updating field %s: %w", fieldName, err)
	}
	return nil
}

// func VerifyVideoIdWithUserId(db *sql.DB, videoID)