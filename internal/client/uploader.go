package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bidhan948/rsync-go/internal/progress"
	"github.com/bidhan948/rsync-go/internal/proto"
)

type UploaderConfig struct {
	ServerURL string
	Token     string
	ChunkSize int64
}

type Uploader struct {
	cfg UploaderConfig
}

func NewUploader(cfg UploaderConfig) *Uploader {
	return &Uploader{cfg: cfg}
}

func (u *Uploader) SendFile(path string, remotePath string) error {
	if remotePath == "" {
		remotePath = filepathBase(path)
	}
	plan, err := planChunks(path, u.cfg.ChunkSize)
	if err != nil {
		return err
	}
	received, err := u.getReceived(remotePath)
	if err != nil {
		received = map[int64]bool{}
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	tracker := progress.NewTracker(plan.Total)
	defer tracker.Done()
	for i := int64(0); i < plan.Count; i++ {
		if received[i] {
			start, end := chunkRange(i, plan.ChunkSize, plan.Total)
			tracker.Add(end - start)
			continue
		}
		start, end := chunkRange(i, plan.ChunkSize, plan.Total)
		bufSize := end - start
		buf := make([]byte, bufSize)
		_, err := f.ReadAt(buf, start)
		if err != nil && err != io.EOF {
			return err
		}
		err = u.sendChunk(remotePath, i, plan.Count, buf)
		if err != nil {
			return err
		}
		tracker.Add(bufSize)
	}
	return nil
}

func (u *Uploader) sendChunk(remotePath string, index int64, totalChunks int64, data []byte) error {
	url := fmt.Sprintf("%s/upload/chunk?path=%s&index=%d&total=%d", trimSlash(u.cfg.ServerURL), remotePath, index, totalChunks)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	if u.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.cfg.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (u *Uploader) getReceived(remotePath string) (map[int64]bool, error) {
	url := fmt.Sprintf("%s/upload/status?path=%s", trimSlash(u.cfg.ServerURL), remotePath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if u.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.cfg.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return map[int64]bool{}, nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status error %d: %s", resp.StatusCode, string(body))
	}
	var st proto.ChunkStatus
	err = json.NewDecoder(resp.Body).Decode(&st)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]bool)
	for _, v := range st.Received {
		m[v] = true
	}
	return m, nil
}

func (u *Uploader) GetRemoteMeta(remotePath string) (*proto.FileMeta, error) {
	url := fmt.Sprintf("%s/files/meta?path=%s", trimSlash(u.cfg.ServerURL), remotePath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if u.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.cfg.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status error %d: %s", resp.StatusCode, string(body))
	}
	var meta proto.FileMeta
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func trimSlash(s string) string {
	if s == "" {
		return s
	}
	if s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}

func filepathBase(path string) string {
	i := len(path) - 1
	for i >= 0 && (path[i] == '/' || path[i] == '\\') {
		i--
	}
	if i < 0 {
		return ""
	}
	j := i
	for j >= 0 && path[j] != '/' && path[j] != '\\' {
		j--
	}
	return path[j+1 : i+1]
}
