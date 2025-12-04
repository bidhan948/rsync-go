package client

import (
	"math"
	"os"
)

type chunkPlan struct {
	ChunkSize int64
	Total     int64
	Count     int64
}

func planChunks(path string, chunkSize int64) (chunkPlan, error) {
	info, err := os.Stat(path)
	if err != nil {
		return chunkPlan{}, err
	}
	size := info.Size()
	if chunkSize <= 0 {
		chunkSize = 4 * 1024 * 1024
	}
	count := int64(math.Ceil(float64(size) / float64(chunkSize)))
	return chunkPlan{
		ChunkSize: chunkSize,
		Total:     size,
		Count:     count,
	}, nil
}

func chunkRange(chunkIndex int64, chunkSize int64, totalSize int64) (int64, int64) {
	start := chunkIndex * chunkSize
	end := start + chunkSize
	if end > totalSize {
		end = totalSize
	}
	if start >= totalSize {
		return 0, 0
	}
	return start, end
}
