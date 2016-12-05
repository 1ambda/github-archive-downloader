package archive

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

        log "github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

// Download github archive files
func Download(dryrun bool, start time.Time, end time.Time) {
	var err error

	for start.Before(end) {
		start = start.Add(time.Hour)

		filename := fmt.Sprintf("%d-%02d-%02d-%02d.json",
			start.Year(), start.Month(), start.Day(), start.Hour())
		url := fmt.Sprintf("http://data.githubarchive.org/%s.gz", filename)

                err = nil
                if !dryrun {
                        err = getGzipJsonAndWriteToFile(url, filename)
                }
                context := log.Ctx{"filename": filename}
		if err != nil {
                        log.Error(err.Error(), context)
                        continue
		}
                log.Info("Downloaded", context)
	}
}

func getGzipJsonAndWriteToFile(url string, filename string) error {
	// 1. Get json
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	if res.StatusCode == 404 {
		err = errors.Errorf("Failed to get %s (404)", url)
		return err
	}

	// 2. Unpack gzip concurrently
	pipeRedaer, pipeWriter := io.Pipe()

	go func() {
		gzipReader, _ := gzip.NewReader(res.Body)

		defer func() {
			res.Body.Close()
			gzipReader.Close()
			pipeWriter.Close()
		}()

		io.Copy(pipeWriter, gzipReader)
	}()

	// 3. Write to file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, pipeRedaer)

	return err
}
