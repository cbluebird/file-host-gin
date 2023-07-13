package models

import "time"

type NameMap struct {
	UUID          string    `json:"uuid"`
	ID            int       `gorm:"primary_key;AUTO_INCREMENT"`
	Date          time.Time `json:"date"`
	Src           string    `json:"src"`
	Path          string    `json:"path"`
	Size          int64     `json:"size"`
	Type          string    `json:"type"`
	DownloadCount int64     `json:"download_count"`
}
