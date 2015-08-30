package download

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/anaminus/rbxplore/event"
)

var ErrClosed = errors.New("download is closed")

type ErrStatus int

func (err ErrStatus) Error() string {
	return strconv.Itoa(int(err)) + " " + http.StatusText(int(err))
}

// Download implements a reusable io.ReadCloser to download data from a URL,
// without any extra setup.
type Download struct {
	// Name is used to identify the download.
	Name string

	// URL is the location to download data from.
	URL string

	// UpdateRate limits how often OnProgress is fired while downloading.
	UpdateRate time.Duration

	onProgress *event.Event
	reader     io.ReadCloser
	lastUpdate time.Time
	progress   int64
	total      int64
	mutex      sync.Mutex
}

// OnProgress is an event that is fired when progress has been made on the
// download. Four arguments are passed to the listener:
//
// `name`: A string identifying the item being downloaded.
//
// `progress`: An int64 indicating the amount of bytes downloaded so far.
//
// `total`: An int64 indicating the total size of the download. This may be
// -1, indicating an unknown size.
//
// `err`: An error, if one has occurred. If the download is finished, the
// error will be io.EOF. If the download was closed before it was finished,
// the error will be ErrClosed.
func (dl *Download) OnProgress(listener func(...interface{})) *event.Connection {
	if dl.onProgress == nil {
		dl.onProgress = new(event.Event)
	}
	return dl.onProgress.Connect(listener)
}

// Read reads a part of the download. In the first read, the download is
// started automatically. If an unexpected HTTP status was encountered, then
// it will be returned as an ErrStatus. When all of the data has been read,
// the downloader is automatically closed, and may then be reused.
func (dl *Download) Read(p []byte) (int, error) {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()

	if dl.reader == nil {
		resp, err := http.Get(dl.URL)
		dl.progress = 0
		if resp != nil {
			dl.total = resp.ContentLength
			if err == nil && resp.StatusCode < 200 || resp.StatusCode >= 300 {
				err = ErrStatus(resp.StatusCode)
			}
		} else {
			dl.total = -1
		}
		dl.lastUpdate = time.Now()
		dl.onProgress.Fire(dl.Name, dl.progress, dl.total, err)
		if err != nil {
			return 0, err
		}
		dl.reader = resp.Body
	}

	n, err := dl.reader.Read(p)
	dl.progress += int64(n)

	if err != nil {
		dl.lastUpdate = time.Now()
		dl.onProgress.Fire(dl.Name, dl.progress, dl.total, err)
		dl.reader.Close()
		dl.reader = nil
	} else {
		t := time.Now()
		if t.Sub(dl.lastUpdate).Seconds() >= dl.UpdateRate.Seconds() {
			dl.lastUpdate = t
			dl.onProgress.Fire(dl.Name, dl.progress, dl.total, nil)
		}
	}

	return n, err
}

// Close can be used to stop the download before it is finished. The download
// may then be reused.
func (dl *Download) Close() error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()

	if dl.reader == nil {
		return ErrClosed
	}
	dl.lastUpdate = time.Now()
	dl.onProgress.Fire(dl.Name, dl.progress, dl.total, ErrClosed)
	err := dl.reader.Close()
	return err
}
