package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appname   = "datasim"
	envPrefix = "DS"

	bucket = "snowflake-workshop-lab"
	region = "us-east-1"
	dsn    = "host=localhost port=5432 user=pguser password=password dbname=db sslmode=disable"

	tripsPrefix     = "citibike-trips"
	tripsFileCap    = 14
	tripsRecordCap  = 50
	tripsWorkerCap  = 3
	ridersPrefix    = "citibike-trips-json"
	ridersFileCap   = 2
	ridersRecordCap = 50
	ridersWorkerCap = 2
	randLowerDelay  = 500
	randUpperDelay  = 7500
	randLowerRecord = 10
	randUpperRecord = 250
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "program exited due to %v\n", err)
		os.Exit(1)
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
	cfg := cliConfig{}

	// Sub command: start.
	cmd := cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("Start simulation"),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectDB(cfg.DSN)
			if err != nil {
				return err
			}

			if err := initDB(db); err != nil {
				return err
			}

			if err := listS3Bucket(cfg.TripConfig(), db, downloadTripData); err != nil {
				return err
			}

			if err := listS3Bucket(cfg.RiderConfig(), db, downloadRiderData); err != nil {
				return err
			}

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
	flags.Int("lower-delay", randLowerDelay, "Lower bound of the random delay (ms) between load batches.")
	flags.Int("upper-delay", randUpperDelay, "Upper bound of the random delay (ms) between load batches.")
	flags.Int("lower-record", randLowerRecord, "Lower bound of the random size of block of records to load.")
	flags.Int("upper-record", randUpperRecord, "Upper bound of the random size of block of records to load.")
	flags.String("trips.prefix", tripsPrefix, "Prefix of the trip data files.")
	flags.Int("trips.file-cap", tripsFileCap, "Number of trip data files to load.")
	flags.Int("trips.record-cap", tripsRecordCap, "Number of trip data records to load.")
	flags.Int("trips.worker-cap", tripsWorkerCap, "Number of concurrent workers for the trip data loading.")
	flags.String("riders.prefix", ridersPrefix, "Prefix of the rider data files.")
	flags.Int("riders.file-cap", ridersFileCap, "Number of rider data files to load.")
	flags.Int("riders.record-cap", ridersRecordCap, "Number of rider data records to load.")
	flags.Int("riders.worker-cap", ridersWorkerCap, "Number of concurrent workers for the rider data loading.")

	err := v.BindPFlags(flags)
	checkErr(err)
	err = flags.Parse(os.Args)
	checkErr(err)

	// Parse.
	err = v.Unmarshal(&cfg)
	checkErr(err)

	return &cmd
}
