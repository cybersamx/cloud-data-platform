package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	bucket       = "snowflake-workshop-lab"
	prefixRiders = "citibike-trips-json"
	prefixTrips  = "citibike-trips"
	region       = "us-east-1"
)

type s3FileHandleFunc func(sess *session.Session, obj *s3.Object, cfg s3Config, db *sql.DB) error
type s3ParseObjectFunc func(reader io.Reader, db *sql.DB, lineCap int) error

type s3Config struct {
	bucket  string
	prefix  string
	region  string
	fileCap int
	lineCap int
}

type trip struct {
	TripDuration     *int
	StartTime        *time.Time
	EndTime          *time.Time
	StartStationID   *int
	StartStationName *string
	StartStationLat  *float32
	StartStationLong *float32
}

// --- Helper Functions ---

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

func listS3Bucket(cfg s3Config, db *sql.DB, handler s3FileHandleFunc) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.region),
	})
	if err != nil {
		return err
	}

	client := s3.New(sess)

	objs, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.bucket),
		Prefix: aws.String(cfg.prefix),
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

	nsize, unit := normalizeSize(size)
	log.Printf("Number of files found: %d", count)
	log.Printf("Total size of files: %.3f %s", nsize, unit)

	// For each file, spawn off a worker to process and load the file.
	numDir := 0

	var wg sync.WaitGroup

	// Make a buffered channel to limit the number of concurrent goroutines.
	workersChan := make(chan struct{}, 4)

	for i, obj := range objs.Contents {
		if i-numDir >= cfg.fileCap {
			break
		}

		if isS3Dir(obj) {
			numDir++
			continue
		}

		if handler == nil {
			if i == 0 {
				log.Printf("List of s3 bucket %s:", cfg.bucket)
			}
			log.Printf("File %s of size %d\n", *obj.Key, *obj.Size)

			continue
		}

		wg.Add(1)
		go func(o *s3.Object) {
			defer wg.Done()

			// If channel is filled to the cap, the goroutine will be blocked until objects
			// are released from the channel.
			workersChan <- struct{}{}

			err := handler(sess, o, cfg, db)
			if err != nil {
				log.Println(err)
			}

			// Consume the channel to release the object.
			<-workersChan
		}(obj)
	}

	wg.Wait()

	return nil
}

func downloadS3Object(sess *session.Session, obj *s3.Object, cfg s3Config, db *sql.DB, handler s3ParseObjectFunc) error {
	log.Printf("Extracting file %s of size %d from s3\n", *obj.Key, *obj.Size)

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
		if err := handler(reader, db, cfg.lineCap); err != nil {
			return err
		}
	}

	return nil
}

// --- Application Specific Functions ---

// downloadRiderData implements the s3FileHandleFunc function type by reading a gzipped, json-formatted trip file,
// parsing the content, and insert the content to postgres.
func downloadRiderData(sess *session.Session, obj *s3.Object, cfg s3Config, db *sql.DB) error {
	parseHandler := func(reader io.Reader, db *sql.DB, lineCap int) error {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		// Read the decompressed content line by line.
		// NOTE: Each line is json encoded string.
		ungzip := bufio.NewScanner(gzReader)
		buf := make([]byte, 0, bufio.MaxScanTokenSize)
		ungzip.Buffer(buf, bufio.MaxScanTokenSize*4)

		n := 0
		for ungzip.Scan() {
			if n >= lineCap {
				return nil
			}

			n++

			line := ungzip.Bytes()
			if ungzip.Err() != nil {
				return ungzip.Err()
			}

			var record map[string]any
			var syntaxErr *json.SyntaxError
			err := json.Unmarshal(line, &record)
			switch {
			case errors.As(err, &syntaxErr):
				log.Printf("Can't parse string to json; err=%v", err)
				log.Printf("String: %s", ungzip.Text())
				continue // Ignore
			case err != nil:
				return err
			}

			log.Println("Inserting:", record)

			stmt := stmtBuilder().
				Insert("riders").
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

	return downloadS3Object(sess, obj, cfg, db, parseHandler)
}

// downloadTripData implements the s3FileHandleFunc function type by reading a gzipped, csv-formatted trip file,
// parsing the content, and insert the content to postgres.
func downloadTripData(sess *session.Session, obj *s3.Object, cfg s3Config, db *sql.DB) error {
	parseHandler := func(reader io.Reader, db *sql.DB, lineCap int) error {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		// Read the decompressed content line by line.
		ungzip := bufio.NewScanner(gzReader)
		buf := make([]byte, 0, bufio.MaxScanTokenSize)
		ungzip.Buffer(buf, bufio.MaxScanTokenSize*4)

		n := 0
		for ungzip.Scan() {
			if n >= lineCap {
				return nil
			}

			n++

			line := ungzip.Bytes()
			if ungzip.Err() != nil {
				return ungzip.Err()
			}

			lineReader := bytes.NewReader(line)
			csvReader := csv.NewReader(lineReader)
			csvReader.FieldsPerRecord = 16
			record, err := csvReader.Read()
			switch {
			case err == io.EOF:
				continue
			case err != nil:
				return err
			}

			log.Println("Inserting:", record)

			stmt := stmtBuilder().
				Insert("trips").
				Columns(
					"trip_duration", "start_time", "stop_time",
					"start_station_id", "start_station_name", "start_station_latitude", "start_station_longitude",
					"end_station_id", "end_station_name", "end_station_latitude", "end_station_longitude",
					"bike_id", "membership_type", "usertype", "birth_year", "gender",
				).
				Values(
					record[0], record[1], record[2],
					record[3], record[4], record[5], record[6],
					record[7], record[8], record[9], record[10],
					record[11], record[12], record[13], record[14], record[15],
				)

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

	return downloadS3Object(sess, obj, cfg, db, parseHandler)
}

func riderDataFromS3(db *sql.DB) error {
	cfg := s3Config{
		bucket:  bucket,
		prefix:  prefixRiders,
		region:  region,
		fileCap: 2,
		lineCap: 3,
	}

	if err := listS3Bucket(cfg, db, downloadRiderData); err != nil {
		return err
	}

	return nil
}

func tripDataFromS3(db *sql.DB) error {
	cfg := s3Config{
		bucket:  bucket,
		prefix:  prefixTrips,
		region:  region,
		fileCap: 14,
		lineCap: 2,
	}

	if err := listS3Bucket(cfg, db, downloadTripData); err != nil {
		return err
	}

	return nil
}
