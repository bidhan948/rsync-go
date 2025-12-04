package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bidhan948/rsync-go/internal/fsutil"
	"github.com/bidhan948/rsync-go/internal/proto"
)

type Config struct {
	Port      int
	BaseDir   string
	AuthToken string
}

func Run(cfg Config) error {
	err := os.MkdirAll(cfg.BaseDir, 0o755)
	if err != nil {
		return err
	}
	cs := newChunkStore(cfg.BaseDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/files/meta", func(w http.ResponseWriter, r *http.Request) {
		if !checkToken(cfg.AuthToken, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}
		full := filepath.Join(cfg.BaseDir, filepath.FromSlash(path))
		info, err := os.Stat(full)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "stat error", http.StatusInternalServerError)
			return
		}
		hash, err := fsutil.HashFile(full)
		if err != nil {
			http.Error(w, "hash error", http.StatusInternalServerError)
			return
		}
		meta := proto.FileMeta{
			Size:    info.Size(),
			ModTime: info.ModTime().UTC(),
			Hash:    hash,
		}
		writeJSON(w, meta)
	})
	mux.HandleFunc("/upload/status", func(w http.ResponseWriter, r *http.Request) {
		if !checkToken(cfg.AuthToken, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}
		received, err := cs.getStatus(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, proto.ChunkStatus{Received: received})
	})
	mux.HandleFunc("/upload/chunk", func(w http.ResponseWriter, r *http.Request) {
		if !checkToken(cfg.AuthToken, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := r.URL.Query().Get("path")
		indexStr := r.URL.Query().Get("index")
		totalStr := r.URL.Query().Get("total")
		if path == "" || indexStr == "" || totalStr == "" {
			http.Error(w, "missing params", http.StatusBadRequest)
			return
		}
		index, err := strconv.ParseInt(indexStr, 10, 64)
		if err != nil {
			http.Error(w, "bad index", http.StatusBadRequest)
			return
		}
		total, err := strconv.ParseInt(totalStr, 10, 64)
		if err != nil {
			http.Error(w, "bad total", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		done, err := cs.addChunk(path, index, total, body)
		if err != nil {
			http.Error(w, "store error", http.StatusInternalServerError)
			return
		}
		if done {
			dest := filepath.Join(cfg.BaseDir, filepath.FromSlash(path))
			err = cs.assemble(path, dest)
			if err != nil {
				http.Error(w, "assemble error", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	addr := ":" + strconv.Itoa(cfg.Port)
	s := &http.Server{
		Addr:         addr,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
	log.Printf("listening on %s baseDir=%s", addr, cfg.BaseDir)
	return s.ListenAndServe()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func checkToken(token string, r *http.Request) bool {
	if token == "" {
		return true
	}
	h := r.Header.Get("Authorization")
	if h == "" {
		return false
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return false
	}
	if parts[0] != "Bearer" {
		return false
	}
	return parts[1] == token
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(v)
}
