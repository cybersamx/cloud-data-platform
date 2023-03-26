package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

type s3ParseObjectFunc func(reader io.Reader, db *sql.DB, tableCfg tableConfig, logger *logrus.Logger) error

// --- Helper Functions ---

func toAnySlice[T any](src []T) []any {
	target := make([]any, len(src), cap(src))
	for i, v := range src {
		target[i] = v
	}

	return target
}

func isPrefix(obj *s3.Object) bool {
	parts := strings.Split(*obj.Key, "/")
	if len(parts) == 0 {
		return false
	}

	return parts[len(parts)-1] == ""
}

var units = [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB"}

func normalizeSize(size int64) (float64, string) {
	fsize := float64(size)

	step := 0
	for fsize < 0.0 || fsize >= 1000.0 {
		fsize /= 1000.0
		step++
	}

	return fsize, units[step]
}

// parseJSON implements the s3FileHandleFunc function type by reading a gzipped, json-formatted trip file,
// parsing the content, and insert the content to postgres.
func parseJSON(reader io.Reader, db *sql.DB, tableCfg tableConfig, logger *logrus.Logger) error {
	if tableCfg.Source.IsGZip {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		reader = gzReader
	}

	// Read the decompressed content line by line.
	// NOTE: Each line is json encoded string.
	ungzip := bufio.NewScanner(reader)
	buf := make([]byte, 0, bufio.MaxScanTokenSize)
	ungzip.Buffer(buf, bufio.MaxScanTokenSize*4)

	lineNum := 0

	for ungzip.Scan() {
		if lineNum >= tableCfg.Source.RowsExtract {
			return nil
		}

		lineNum++

		line := ungzip.Bytes()
		if ungzip.Err() != nil {
			return ungzip.Err()
		}

		var record map[string]any
		var syntaxErr *json.SyntaxError
		err := json.Unmarshal(line, &record)
		switch {
		case errors.As(err, &syntaxErr):
			logger.WithError(err).Errorf("Can't parse string to json; string value: %q", ungzip.Text())
			continue // Ignore
		case err != nil:
			return err
		}

		logger.Traceln("Inserting:", record)

		stmt := stmtBuilder().
			Insert(tableCfg.Name).
			Columns("json").
			Values(string(line))

		query, args, err := stmt.ToSql()
		if err != nil {
			return err
		}

		_, err = db.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseCSV implements the s3FileHandleFunc function type by reading a gzipped, csv-formatted trip file,
// parsing the content, and insert the content to postgres.
func parseCSV(reader io.Reader, db *sql.DB, tableCfg tableConfig, logger *logrus.Logger) error {
	if tableCfg.Source.IsGZip {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		reader = gzReader
	}

	// Read the decompressed content line by line.
	ungzip := bufio.NewScanner(reader)
	buf := make([]byte, 0, bufio.MaxScanTokenSize)
	ungzip.Buffer(buf, bufio.MaxScanTokenSize*4)

	lineNum := 0

	for ungzip.Scan() {
		if lineNum >= tableCfg.Source.RowsExtract {
			return nil
		}

		lineNum++

		line := ungzip.Bytes()
		if ungzip.Err() != nil {
			return ungzip.Err()
		}

		lineReader := bytes.NewReader(line)
		csvReader := csv.NewReader(lineReader)
		csvReader.FieldsPerRecord = len(tableCfg.Columns)
		record, err := csvReader.Read()
		switch {
		case err == io.EOF:
			continue
		case err != nil:
			return err
		}

		logger.Traceln("Inserting:", record)

		var columns []string
		for _, col := range tableCfg.Columns {
			columns = append(columns, col.Name)
		}

		stmt := stmtBuilder().
			Insert(tableCfg.Name).
			Columns(columns...).
			Values(toAnySlice[string](record)...)

		query, args, err := stmt.ToSql()
		if err != nil {
			return err
		}

		_, err = db.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// --- S3Connector ---

type s3Connector struct {
	db        *sql.DB
	logger    *logrus.Logger
	connCfg   connConfig
	sess      *session.Session
	client    *s3.S3
	procFiles int
	mu        sync.Mutex
}

func newS3Connector(db *sql.DB, logger *logrus.Logger, connCfg connConfig) (*s3Connector, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(connCfg.Region),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		return nil, err
	}

	conn := s3Connector{
		db:      db,
		logger:  logger,
		connCfg: connCfg,
		sess:    sess,
		client:  s3.New(sess),
	}

	return &conn, nil
}

func (sc *s3Connector) incProcFiles(n int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.procFiles += n
}

func (sc *s3Connector) listS3Prefix(tableCfg tableConfig, prefix string, handler s3ParseObjectFunc) error {
	inputs := s3.ListObjectsV2Input{
		Bucket:    aws.String(sc.connCfg.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	err := sc.client.ListObjectsV2Pages(&inputs, func(page *s3.ListObjectsV2Output, isLastPage bool) bool {
		if tableCfg.Source.Recursive {
			for _, p := range page.CommonPrefixes {
				if err := sc.listS3Prefix(tableCfg, *p.Prefix, handler); err != nil {
					return false
				}

				if sc.procFiles > tableCfg.Source.FilesExtract {
					return false
				}
			}
		}

		// For each file, spawn off a worker to process and load the file.
		var wg sync.WaitGroup

		// Make a buffered channel to limit the number of concurrent goroutines.
		workersChan := make(chan struct{}, tableCfg.Source.Workers)

		for _, obj := range page.Contents {
			if isPrefix(obj) {
				continue
			}

			sc.procFiles++

			if sc.procFiles > tableCfg.Source.FilesExtract {
				break
			}

			wg.Add(1)
			go func(o *s3.Object) {
				defer wg.Done()

				// If channel is filled to the cap, the goroutine will be blocked until objects
				// are released from the channel.
				workersChan <- struct{}{}

				err := sc.downloadS3Object(o, sc.connCfg.Bucket, tableCfg, handler)
				if err != nil {
					sc.logger.WithError(err)
				}

				// Delay for sometime before running the next handler.
				sc.logger.Infof("Done extracting %s. Sleep for %v before the next extract",
					*o.Key, tableCfg.Source.NextExtractDelay)
				time.Sleep(tableCfg.Source.NextExtractDelay)

				// Consume the channel to release the object.
				<-workersChan
			}(obj)
		}

		wg.Wait()

		return sc.procFiles <= tableCfg.Source.FilesExtract
	})

	return err
}

func (sc *s3Connector) listS3Bucket(tableCfg tableConfig, handler s3ParseObjectFunc) error {
	if handler == nil {
		return errors.New("listS3Bucket function requires a handler s3ParseObjectFunc")
	}

	if err := sc.listS3Prefix(tableCfg, tableCfg.Source.Prefix, handler); err != nil {
		return err
	}

	return nil
}

func (sc *s3Connector) downloadS3Object(obj *s3.Object, bucket string, tableCfg tableConfig, handler s3ParseObjectFunc) error {
	fsize, funit := normalizeSize(*obj.Size)
	sc.logger.WithFields(logrus.Fields{"size": fmt.Sprintf("%.3f %s", fsize, funit)}).
		Infof("Extracting top %d records from file %s:", tableCfg.Source.RowsExtract, *obj.Key)

	downloader := s3manager.NewDownloader(sc.sess)
	downloader.Concurrency = 3

	// This isn't optimal. Reuses the buffer if it's big enough to load the entire file,
	// otherwise creates a new buffer.
	buf := make([]byte, *obj.Size)

	// Download the file.
	w := aws.NewWriteAtBuffer(buf)
	_, err := downloader.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(*obj.Key),
	})
	if err != nil {
		return err
	}

	if handler != nil {
		reader := bytes.NewReader(buf)
		if err := handler(reader, sc.db, tableCfg, sc.logger); err != nil {
			return err
		}
	}

	return nil
}
