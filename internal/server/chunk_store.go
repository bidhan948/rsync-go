package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type chunkStore struct {
	baseDir string
}

type chunkMeta struct {
	Total    int64   `json:"total"`
	Received []int64 `json:"received"`
}

func newChunkStore(baseDir string) *chunkStore {
	return &chunkStore{baseDir: baseDir}
}

func (c *chunkStore) chunkDir(path string) string {
	return filepath.Join(c.baseDir, ".chunks", path)
}

func (c *chunkStore) metaPath(path string) string {
	return filepath.Join(c.chunkDir(path), "meta.json")
}

func (c *chunkStore) chunkPath(path string, index int64) string {
	return filepath.Join(c.chunkDir(path), fmt.Sprintf("%d.part", index))
}

func (c *chunkStore) loadMeta(path string) (chunkMeta, error) {
	m := chunkMeta{}
	p := c.metaPath(path)
	data, err := os.ReadFile(p)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func (c *chunkStore) saveMeta(path string, m chunkMeta) error {
	dir := c.chunkDir(path)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.metaPath(path), data, 0o644)
}

func (c *chunkStore) addChunk(path string, index int64, total int64, data []byte) (bool, error) {
	dir := c.chunkDir(path)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return false, err
	}
	chunkFile := c.chunkPath(path, index)
	err = os.WriteFile(chunkFile, data, 0o644)
	if err != nil {
		return false, err
	}
	m := chunkMeta{}
	existing, err := c.loadMeta(path)
	if err == nil {
		m = existing
	}
	m.Total = total
	seen := make(map[int64]bool)
	for _, v := range m.Received {
		seen[v] = true
	}
	if !seen[index] {
		m.Received = append(m.Received, index)
	}
	err = c.saveMeta(path, m)
	if err != nil {
		return false, err
	}
	if int64(len(m.Received)) == m.Total {
		return true, nil
	}
	return false, nil
}

func (c *chunkStore) getStatus(path string) ([]int64, error) {
	m, err := c.loadMeta(path)
	if err != nil {
		return nil, err
	}
	return m.Received, nil
}

func (c *chunkStore) assemble(path string, dest string) error {
	m, err := c.loadMeta(path)
	if err != nil {
		return err
	}
	outDir := filepath.Dir(dest)
	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	for i := int64(0); i < m.Total; i++ {
		partPath := c.chunkPath(path, i)
		data, err := os.ReadFile(partPath)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
