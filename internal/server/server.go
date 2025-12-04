package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bidhan948/rsync-go/internal/ui"
)

type Config struct {
	ListenAddr string
	OutputDir  string
}

func Run(cfg Config) error {
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB memory buffer
			http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "get file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename := filepath.Base(header.Filename)
		dstPath := filepath.Join(cfg.OutputDir, filename)

		ui.Info("Receiving file: " + filename)

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "create dst: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		written, err := io.Copy(dst, file)
		if err != nil {
			http.Error(w, "write file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ui.Success(fmt.Sprintf("Saved %s (%d bytes)", filename, written))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	ui.Success("Server listening on " + cfg.ListenAddr)
	ui.Info("Upload endpoint: http://" + cfg.ListenAddr + "/upload")

	return http.ListenAndServe(cfg.ListenAddr, mux)
}
