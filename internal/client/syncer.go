package client

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/bidhan948/rsync-go/internal/fsutil"
)

type SyncConfig struct {
	ServerURL string
	Token     string
	ChunkSize int64
	Workers   int
	LocalDir  string
	RemoteDir string
}

type Syncer struct {
	cfg      SyncConfig
	uploader *Uploader
}

func NewSyncer(cfg SyncConfig) *Syncer {
	u := NewUploader(UploaderConfig{
		ServerURL: cfg.ServerURL,
		Token:     cfg.Token,
		ChunkSize: cfg.ChunkSize,
	})
	return &Syncer{
		cfg:      cfg,
		uploader: u,
	}
}

func (s *Syncer) Sync() error {
	files, err := fsutil.WalkDir(s.cfg.LocalDir)
	if err != nil {
		return err
	}
	workers := s.cfg.Workers
	if workers <= 0 {
		workers = 4
	}
	wg := sync.WaitGroup{}
	ch := make(chan fsutil.WalkFile)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range ch {
				remotePath := pathJoin(s.cfg.RemoteDir, f.RelPath)
				err := s.syncFile(f, remotePath)
				if err != nil {
					log.Printf("sync error %s: %v", f.RelPath, err)
				}
			}
		}()
	}
	for _, f := range files {
		ch <- f
	}
	close(ch)
	wg.Wait()
	return nil
}

func (s *Syncer) syncFile(f fsutil.WalkFile, remotePath string) error {
	meta, err := s.uploader.GetRemoteMeta(remotePath)
	if err != nil {
		return err
	}
	if meta != nil {
		info, err := os.Stat(f.AbsPath)
		if err != nil {
			return err
		}
		if meta.Size == info.Size() && meta.ModTime.Equal(info.ModTime().UTC()) {
			return nil
		}
	}
	err = s.uploader.SendFile(f.AbsPath, remotePath)
	if err != nil {
		return err
	}
	return nil
}

func pathJoin(base string, rel string) string {
	return filepath.ToSlash(filepath.Join(base, rel))
}
