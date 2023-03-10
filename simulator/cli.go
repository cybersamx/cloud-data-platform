package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appname   = "simulator"
	envPrefix = "DS"

	bucket        = "snowflake-workshop-lab"
	region        = "us-east-1"
	dsn           = "host=localhost port=5433 user=pguser password=password dbname=db sslmode=disable"
	trace         = false
	tripsPrefix   = "citibike-trips"
	ridersPrefix  = "citibike-trips-json"
	filesLoad     = 14
	rowsLoad      = 50
	nextLoadDelay = 5 * time.Minute
	workers       = 2
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "program exited due to %v\n", err)
		os.Exit(1)
	}
}

func newLogger(cfg config) *logrus.Logger {
	logLevel := logrus.InfoLevel
	if cfg.Trace {
		logLevel = logrus.TraceLevel
	}

	return &logrus.Logger{
		Out:   os.Stdout,
		Level: logLevel,
		Formatter: &logrus.TextFormatter{
			ForceColors:            true,
			FullTimestamp:          true,
			TimestampFormat:        "2006-0102 15:04:05",
			DisableLevelTruncation: true,
		},
	}
}

func rootCommand() *cobra.Command {
	// Root command.
	cmd := cobra.Command{
		Use: appname,
		RunE: func(cmd *cobra.Command, args []string) error {
			// User must enter a command, otherwise display the help menu.
			return cmd.Help()
		},
	}

	cmd.AddCommand(serveCommand())

	return &cmd
}

func serveCommand() *cobra.Command {
	cfg := config{}

	// Sub command: start.
	cmd := cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("Start simulation"),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := newLogger(cfg)

			// Database.
			db, err := connectDB(cfg.DSN)
			if err != nil {
				return err
			}

			if err := initDB(db); err != nil {
				return err
			}

			// List and download objects from S3.
			var wg sync.WaitGroup

			prefixes := []string{tripsPrefix, ridersPrefix}
			for _, prefix := range prefixes {
				cfg.Prefix = prefix

				wg.Add(1)
				go func(cfg config) {
					defer wg.Done()

					if err := listS3Bucket(cfg, db, logger, downloadTripData); err != nil {
						logger.WithError(err)
					}
				}(cfg)
			}

			wg.Wait()

			return nil
		},
	}

	v := viper.New()

	// Environment variable.
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// Config file.
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Flags.
	flags := cmd.Flags()
	flags.String("bucket", bucket, "The bucket containing the data files.")
	flags.String("region", region, "The AWS Region associated with the bucket.")
	flags.String("dsn", dsn, "DSN of the database to which the data loads.")
	flags.Bool("trace", trace, "Enable tracing if true.")
	flags.Int("files-load", filesLoad, "Number of data files to load.")
	flags.Int("rows-load", rowsLoad, "Number of rows to load.")
	flags.Duration("next-load-delay", nextLoadDelay, "The delay between loads.")
	flags.Int("workers", workers, "Number of concurrent workers for the data loading.")

	err := v.BindPFlags(flags)
	checkErr(err)
	err = flags.Parse(os.Args)
	checkErr(err)

	if err := v.ReadInConfig(); err != nil {
		checkErr(err)
	}

	// Parse.
	err = v.Unmarshal(&cfg)
	checkErr(err)

	return &cmd
}
