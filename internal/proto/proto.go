package proto

import "time"

type FileMeta struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Hash    string    `json:"hash"`
}

type ChunkStatus struct {
	Received []int64 `json:"received"`
}
