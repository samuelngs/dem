package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Downloader interface
type Downloader interface {
	Start(chan<- int) error
}

type downloader struct {
	URL  string
	Dest string
}

func (v *downloader) Start(progress chan<- int) error {

	out, err := os.Create(v.Dest)
	if err != nil {
		return err
	}
	defer out.Close()

	done := make(chan int64)

	resp, err := http.Get(v.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	go func() {
		for {
			select {
			case <-done:
				break
			default:
				f, err := os.Open(v.Dest)
				if err != nil {
					return
				}
				s, err := f.Stat()
				if err != nil {
					return
				}
				percentage := int(float64(s.Size()) / float64(size) * 100.0)
				progress <- percentage
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return fmt.Errorf("(%d) unable to download package %s", resp.StatusCode, filepath.Base(v.Dest))
	}

	n, err := io.Copy(out, resp.Body)
	done <- n

	return err
}

// New creates a download manager
func New(url, dest string) Downloader {
	return &downloader{
		URL:  url,
		Dest: dest,
	}
}
