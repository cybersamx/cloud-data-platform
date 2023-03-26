package main

import (
	"time"
)

type config struct {
	Conn   connConfig    `mapstructure:"conn"`
	Tables []tableConfig `mapstructure:"tables"`
}

type connConfig struct {
	Bucket string `mapstructure:"bucket"`
	Region string `mapstructure:"region"`
	DSN    string `mapstructure:"dsn"`
	Trace  bool   `mapstructure:"trace"`
}

type tableConfig struct {
	Name    string         `mapstructure:"name"`
	Source  sourceConfig   `mapstructure:"source"`
	Columns []columnConfig `mapstructure:"columns"`
}

type sourceConfig struct {
	Prefix           string        `mapstructure:"prefix"`
	Type             string        `mapstructure:"type"`
	HasHeader        bool          `mapstructure:"has-header"`
	Recursive        bool          `mapstructure:"recursive"`
	IsGZip           bool          `mapstructure:"is-gzip"`
	FilesExtract     int           `mapstructure:"files-extract"`
	RowsExtract      int           `mapstructure:"rows-extract"`
	NextExtractDelay time.Duration `mapstructure:"next-extract-delay"`
	Workers          int           `mapstructure:"workers"`
}

type columnConfig struct {
	Name     string `mapstructure:"name"`
	DataType string `mapstructure:"datatype"`
}
