package main

import (
	"time"
)

type config struct {
	Bucket        string        `mapstructure:"bucket"`
	Region        string        `mapstructure:"region"`
	DSN           string        `mapstructure:"dsn"`
	Trace         bool          `mapstructure:"trace"`
	Prefix        string        `mapstructure:"-"`
	TripsFile     string        `mapstructure:"trips-file"`
	RidersFile    string        `mapstructure:"riders-file"`
	FilesLoad     int           `mapstructure:"files-load"`
	RowsLoad      int           `mapstructure:"rows-load"`
	NextLoadDelay time.Duration `mapstructure:"next-load-delay"`
	Workers       int           `mapstructure:"workers"`
}
