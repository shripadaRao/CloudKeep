package models


type UploadInitializationResponse struct {
	VideoID         string              `json:"video_id"`
    UserID          string              `json:"user_id"`
    ChunkFileNames  []string            `json:"chunk_filenames"`
	CheckSumMap     map[string]string   `json:"check_sum_map"`
    Status          VideoStatus         `json:"Status"`
}

type VideoStatus string

const (
    VideoNotUploaded VideoStatus = "PENDING"
    UploadedToServer VideoStatus = "PROCESSING"
    UploadedToS3     VideoStatus = "COMPLETED"
)

type Video struct {
    VideoID      string      `json:"video_id"`
    UserID       string      `json:"user_id"`
    Status       VideoStatus `json:"status"`
    TotalChunks  int         `json:"total_chunks"`
    CreatedAt    string      `json:"created_at"`
    VideoPath    string      `json:"video_path"`
}

type ChunkStatus string

const (
    ChunkNotUploaded ChunkStatus = "NOT-UPLOADED"
    InServer    ChunkStatus = "IN-SERVER"
    InS3        ChunkStatus = "IN-S3"
)

type VideoChunk struct {
    ChunkID    string      `json:"chunk_id"`
    VideoID    string      `json:"video_id"`
    ChunkNo    int         `json:"chunk_no"`
    Status     ChunkStatus `json:"status"`
    ChunkPath  string      `json:"chunk_path"`
    CheckSum   string      `json:"check_sum"`
    CreatedAt  string      `json:"created_at"`
}

type ChunkVerificationDetails struct {
	ChunkID  string
	VideoID  string
	UserID   string
	Checksum string
}