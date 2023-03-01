package main

type cliConfig struct {
	Bucket      string `mapstructure:"bucket"`
	Region      string `mapstructure:"region"`
	DSN         string `mapstructure:"dsn"`
	LowerDelay  int    `mapstructure:"lower-delay"`
	UpperDelay  int    `mapstructure:"upper-delay"`
	LowerRecord int    `mapstructure:"lower-record"`
	UpperRecord int    `mapstructure:"upper-record"`
	Trips       struct {
		Prefix    string `mapstructure:"prefix"`
		FileCap   int    `mapstructure:"file-cap"`
		RecordCap int    `mapstructure:"record-cap"`
		WorkerCap int    `mapstructure:"worker-cap"`
	} `mapstructure:"trips"`
	Riders struct {
		Prefix    string `mapstructure:"prefix"`
		FileCap   int    `mapstructure:"file-cap"`
		RecordCap int    `mapstructure:"record-cap"`
		WorkerCap int    `mapstructure:"worker-cap"`
	} `mapstructure:"riders"`
}

func (cc cliConfig) RiderConfig() config {
	return config{
		bucket:      cc.Bucket,
		region:      cc.Region,
		prefix:      cc.Riders.Prefix,
		lowerDelay:  cc.LowerDelay,
		upperDelay:  cc.UpperDelay,
		lowerRecord: cc.LowerRecord,
		upperRecord: cc.UpperRecord,
		fileCap:     cc.Riders.FileCap,
		recordCap:   cc.Riders.RecordCap,
		workerCap:   cc.Riders.WorkerCap,
	}
}

func (cc cliConfig) TripConfig() config {
	return config{
		bucket:      cc.Bucket,
		region:      cc.Region,
		prefix:      cc.Trips.Prefix,
		lowerDelay:  cc.LowerDelay,
		upperDelay:  cc.UpperDelay,
		lowerRecord: cc.LowerRecord,
		upperRecord: cc.UpperRecord,
		fileCap:     cc.Trips.FileCap,
		recordCap:   cc.Trips.RecordCap,
		workerCap:   cc.Trips.WorkerCap,
	}
}
