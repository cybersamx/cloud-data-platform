package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

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

type s3ListHandler func(sess *session.Session, obj *s3.Object, cfg s3Config) error
type s3ParseObjectHandler func(reader io.Reader, lineCap int) error

type s3Config struct {
	bucket  string
	prefix  string
	region  string
	fileCap int
	lineCap int
}

// --- Helper Functions ---

func isS3Dir(obj *s3.Object) bool {
	parts := strings.Split(*obj.Key, "/")
	if len(parts) == 0 {
		return false
	}

	return parts[len(parts)-1] == ""
}

func listS3Bucket(cfg s3Config, handler s3ListHandler) error {
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

	numDir := 0
	for i, obj := range objs.Contents {
		if i-numDir >= cfg.fileCap {
			return nil
		}

		if isS3Dir(obj) {
			numDir++
			continue
		}

		if handler == nil {
			if i == 0 {
				log.Printf("List of s3 bucket %s:", cfg.bucket)
			}
			fmt.Printf("File %s of size %d\n", *obj.Key, *obj.Size)

			continue
		}

		if err := handler(sess, obj, cfg); err != nil {
			return err
		}
	}

	return nil
}

func downloadS3Object(sess *session.Session, obj *s3.Object, cfg s3Config, handler s3ParseObjectHandler) error {
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
		if err := handler(reader, cfg.lineCap); err != nil {
			return err
		}
	}

	return nil
}

// --- Application Specific Functions ---

func downloadRiderData(sess *session.Session, obj *s3.Object, cfg s3Config) (rerr error) {
	handler := func(reader io.Reader, lineCap int) error {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer func() {
			rerr = gzReader.Close()
		}()

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

			var dict map[string]any
			var syntaxErr *json.SyntaxError
			err := json.Unmarshal(line, &dict)
			switch {
			case errors.As(err, &syntaxErr):
				log.Printf("Can't parse string to json; err=%v", err)
				log.Printf("String: %s", ungzip.Text())
				continue // Ignore
			case err != nil:
				return err
			}

			fmt.Println(dict)
		}

		return nil
	}

	return downloadS3Object(sess, obj, cfg, handler)
}

func riderDataFromS3() error {
	cfg := s3Config{
		bucket:  bucket,
		prefix:  prefixRiders,
		region:  region,
		fileCap: 1,
		lineCap: 3,
	}

	//var handler s3ListHandler
	handler := downloadRiderData

	if err := listS3Bucket(cfg, handler); err != nil {
		return err
	}

	return nil
}
