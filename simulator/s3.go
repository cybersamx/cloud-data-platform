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

func isS3Dir(obj *s3.Object) bool {
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
	for fsize < 0.0 || fsize > 1000.0 {
		fsize /= 1024.0
		step++
	}

	return fsize, units[step]
}

func listS3Bucket(db *sql.DB, logger *logrus.Logger, connCfg connConfig, tableCfg tableConfig, handler s3ParseObjectFunc) error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(connCfg.Region),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		return err
	}

	client := s3.New(sess)

	objs, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(connCfg.Bucket),
		Prefix: aws.String(tableCfg.Source.Prefix),
	})
	if err != nil {
		return err
	}

	size := int64(0)
	count := 0
	for _, obj := range objs.Contents {
		size += *obj.Size
		count++
	}

	tsize, tunit := normalizeSize(size)
	logger.WithFields(logrus.Fields{"#files": count, "size": fmt.Sprintf("%.3f %s", tsize, tunit)}).
		Infof("Stats for s3://%s/%s", connCfg.Bucket, tableCfg.Source.Prefix)

	numDir := 0

	// For each file, spawn off a worker to process and load the file.
	var wg sync.WaitGroup

	// Make a buffered channel to limit the number of concurrent goroutines.
	workersChan := make(chan struct{}, tableCfg.Source.Workers)

	for i, obj := range objs.Contents {
		if i-numDir >= tableCfg.Source.FilesExtract {
			break
		}

		if isS3Dir(obj) {
			numDir++
			continue
		}

		// If no handler is passed, we just list the bucket content.
		if handler == nil {
			if i == 0 {
				logger.Infof("List of s3://%s/%s:", connCfg.Bucket, tableCfg.Source.Prefix)
			}
			fsize, funit := normalizeSize(*obj.Size)
			logger.WithFields(logrus.Fields{"size": fmt.Sprintf("%.3f %s", fsize, funit)}).
				Infof("File %s:", *obj.Key)

			continue
		}

		wg.Add(1)
		go func(o *s3.Object) {
			defer wg.Done()

			// If channel is filled to the cap, the goroutine will be blocked until objects
			// are released from the channel.
			workersChan <- struct{}{}

			err := downloadS3Object(sess, o, connCfg.Bucket, tableCfg, db, logger, handler)
			if err != nil {
				logger.WithError(err)
			}

			// Delay for sometime before running the next handler.
			logger.Infof("Done extracting %s. Sleep for %v before the next extract",
				*o.Key, tableCfg.Source.NextExtractDelay)
			time.Sleep(tableCfg.Source.NextExtractDelay)

			// Consume the channel to release the object.
			<-workersChan
		}(obj)
	}

	wg.Wait()

	return nil
}

func downloadS3Object(sess *session.Session, obj *s3.Object, bucket string, tableCfg tableConfig, db *sql.DB, logger *logrus.Logger, handler s3ParseObjectFunc) error {
	fsize, funit := normalizeSize(*obj.Size)
	logger.WithFields(logrus.Fields{"size": fmt.Sprintf("%.3f %s", fsize, funit)}).
		Infof("Extracting top %d records from file %s:", tableCfg.Source.RowsExtract, *obj.Key)

	downloader := s3manager.NewDownloader(sess)
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
		if err := handler(reader, db, tableCfg, logger); err != nil {
			return err
		}
	}

	return nil
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
