package models

import "time"

type TokenBucket struct {
    AvailableTokens int    `json:"availableTokens"`
    LastRefilled   time.Time `json:"lastRefilled"`
}
