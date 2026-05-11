package pack

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PackReproducible creates a deterministic .tar.gz of srcDir at dstFile.
// Returns the sha256 digest of the resulting tarball.
func PackReproducible(srcDir, dstFile string, sourceDateEpoch int64) (string, error) {
	var entries []string
	if err := filepath.Walk(srcDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(srcDir, p)
			if err != nil {
				return err
			}
			entries = append(entries, rel)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(entries)

	out, err := os.Create(dstFile)
	if err != nil {
		return "", err
	}
	gz := gzip.NewWriter(out)
	gz.ModTime = time.Unix(0, 0)
	tw := tar.NewWriter(gz)

	mtime := time.Unix(sourceDateEpoch, 0).UTC()

	for _, rel := range entries {
		full := filepath.Join(srcDir, rel)
		info, err := os.Stat(full)
		if err != nil {
			return "", err
		}
		hdr := &tar.Header{
			Name:    rel,
			Mode:    0o644,
			Size:    info.Size(),
			ModTime: mtime,
			Uid:     0,
			Gid:     0,
			Uname:   "",
			Gname:   "",
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", err
		}
		f, err := os.Open(full)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}

	return sha256OfFile(dstFile)
}

func sha256OfFile(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// ReadSourceDateEpoch reads $SOURCE_DATE_EPOCH or returns 0 if unset.
func ReadSourceDateEpoch() int64 {
	v := os.Getenv("SOURCE_DATE_EPOCH")
	if v == "" {
		return 0
	}
	var n int64
	fmt.Sscanf(v, "%d", &n)
	return n
}
