package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	bucket = "snowflake-workshop-lab"
	prefix = "citibike-trips-json"
	region = "us-east-1"
)

var (
	buf []byte
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		panic(err)
	}

	client := s3.New(sess)

	objs, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		panic(err)
	}

	for _, obj := range objs.Contents {
		if isDir(obj) {
			continue
		}

		fmt.Printf("Extracting file %s of size %d from s3\n", *obj.Key, *obj.Size)
		if err := download(sess, obj); err != nil {
			panic(err)
		}
	}
}

func isDir(obj *s3.Object) bool {
	parts := strings.Split(*obj.Key, "/")
	if len(parts) == 0 {
		return false
	}

	return parts[len(parts)-1] == ""
}

func download(sess *session.Session, obj *s3.Object) (rerr error) {
	downloader := s3manager.NewDownloader(sess)
	downloader.Concurrency = 3

	// This isn't optimal. Reuses the buffer if it's big enough to load the entire file,
	// otherwise creates a new buffer.
	fileSize := *obj.Size
	if fileSize > int64(cap(buf)) {
		buf = make([]byte, fileSize, 2*fileSize)
	}

	// Download a gzip file.
	w := aws.NewWriteAtBuffer(buf)
	_, err := downloader.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(*obj.Key),
	})

	// Once the writer has filled the buffer, read off of it and decompress the content.
	reader := bytes.NewReader(buf)
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
	ungzip.Buffer(buf, bufio.MaxScanTokenSize*4)
	for ungzip.Scan() {
		line := ungzip.Bytes()
		if ungzip.Err() != nil {
			return ungzip.Err()
		}

		var dict map[string]any
		if err := json.Unmarshal(line, &dict); err != nil {
			return err
		}

		fmt.Println(dict)
	}

	return nil
}
