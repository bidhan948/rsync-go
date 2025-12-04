package client

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bidhan948/rsync-go/internal/ui"
)

type SendConfig struct {
	FilePath string
	Server   string // e.g. http://192.168.1.20:8080/upload
}

type progressWriter struct {
	total       int64
	sent        int64
	lastPrinted time.Time
	start       time.Time
	filename    string
	barWidth    int
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.sent += int64(n)

	now := time.Now()
	if p.lastPrinted.IsZero() || now.Sub(p.lastPrinted) > 150*time.Millisecond {
		p.printProgress(now)
		p.lastPrinted = now
	}
	return n, nil
}

func (p *progressWriter) printProgress(now time.Time) {
	percent := float64(p.sent) / float64(p.total) * 100
	if percent > 100 {
		percent = 100
	}
	elapsed := now.Sub(p.start).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}
	speed := float64(p.sent) / 1024.0 / 1024.0 / elapsed // MB/s

	// progress bar
	filled := int(percent / 100 * float64(p.barWidth))
	if filled > p.barWidth {
		filled = p.barWidth
	}
	bar := ""
	for i := 0; i < p.barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "─"
		}
	}

	fmt.Printf("\r%s %.1f%% |%s| %.2f MB/s (%d / %d bytes)",
		p.filename,
		percent,
		bar,
		speed,
		p.sent,
		p.total,
	)
}

func SendFile(cfg SendConfig) error {
	file, err := os.Open(cfg.FilePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, expected a file")
	}

	ui.Info(fmt.Sprintf("File size: %d bytes", info.Size()))

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	progress := &progressWriter{
		total:    info.Size(),
		filename: filepath.Base(cfg.FilePath),
		start:    time.Now(),
		barWidth: 40,
	}

	// Goroutine: write multipart form (streaming) into the pipe
	go func() {
		defer pw.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(cfg.FilePath))
		if err != nil {
			ui.Error("create form file: " + err.Error())
			return
		}

		multi := io.MultiWriter(part, progress)

		if _, err := io.Copy(multi, file); err != nil {
			ui.Error("copy file: " + err.Error())
			return
		}

		// Finish progress bar at 100%
		progress.printProgress(time.Now())
		fmt.Println()

		if err := writer.Close(); err != nil {
			ui.Error("writer close: " + err.Error())
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, cfg.Server, pr)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 0, // no timeout for large files; you can tune it
	}

	ui.Info("Uploading...")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %s: %s", resp.Status, string(body))
	}

	return nil
}
