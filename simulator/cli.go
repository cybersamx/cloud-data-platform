package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appname   = "simulator"
	envPrefix = "DS"

	bucket = "snowflake-workshop-lab"
	region = "us-east-1"
	dsn    = "host=localhost port=5433 user=postgres password=password dbname=db sslmode=disable"
	trace  = false
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "program exited due to %v\n", err)
		os.Exit(1)
	}
}

func newLogger(cfg config) *logrus.Logger {
	logLevel := logrus.InfoLevel
	if cfg.Conn.Trace {
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

	cmd.AddCommand(startCommand())

	return &cmd
}

func startCommand() *cobra.Command {
	cfg := config{}

	// Sub command: start.
	cmd := cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("Start simulation"),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := newLogger(cfg)

			// Database.
			db, err := connectDB(cfg.Conn.DSN)
			if err != nil {
				return err
			}

			if err := initDB(db, cfg); err != nil {
				return err
			}

			// List and download objects from S3.
			var wg sync.WaitGroup

			for _, tableCfg := range cfg.Tables {
				wg.Add(1)

				go func(connCfg connConfig, tableCfg tableConfig) {
					defer wg.Done()

					var handler s3ParseObjectFunc
					switch tableCfg.Source.Type {
					case "csv":
						handler = parseCSV
					case "json":
						handler = parseJSON
					default:
						panic(errors.New("missing or invalid table source type"))
					}

					sc, err := newS3Connector(db, logger, cfg.Conn)
					if err != nil {
						logger.WithError(err)
						return
					}

					if err := sc.listS3Bucket(tableCfg, handler); err != nil {
						logger.WithError(err)
						return
					}
				}(cfg.Conn, tableCfg)
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

	// Flags - the tables field is only set by the config file.
	flags := cmd.Flags()
	flags.String("conn.bucket", bucket, "The bucket containing the data files.")
	flags.String("conn.region", region, "The AWS Region associated with the bucket.")
	flags.String("conn.dsn", dsn, "DSN of the database to which the data loads.")
	flags.Bool("conn.trace", trace, "Enable tracing if true.")

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
