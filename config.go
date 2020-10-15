package ytcompare

const (
	//MongoDBURLField field name of mongodb-url
	MongoDBURLField = "mongodb-url"
	//DBNameField field name of db-name
	DBNameField = "db-name"

	//AllSyncURLsField Field name of all-sync-urls
	AllSyncURLsField = "all-sync-urls"
	//StartTimeField Field name of start-time
	StartTimeField = "start-time"
	//TimeRangeField Field name of time-range
	TimeRangeField = "time-range"
	//WaitTimeField Field name of wait-time
	WaitTimeField = "wait-time"
	//SkipTimeField Field name of skip-time
	SkipTimeField = "skip-time"

	//LoggerOutputField Field name of logger.output config
	LoggerOutputField = "logger.output"
	//LoggerFilePathField Field name of logger.file-path config
	LoggerFilePathField = "logger.file-path"
	//LoggerRotationTimeField Field name of logger.rotation-time config
	LoggerRotationTimeField = "logger.rotation-time"
	//LoggerMaxAgeField Field name of logger.rotation-time config
	LoggerMaxAgeField = "logger.max-age"
	//LoggerLevelField Field name of logger.level config
	LoggerLevelField = "logger.level"
)

//Config system configuration
type Config struct {
	MongoDBURL  string     `mapstructure:"mongodb-url"`
	DBName      string     `mapstructure:"db-name"`
	AllSyncURLs []string   `mapstructure:"all-sync-urls"`
	StartTime   int        `mapstructure:"start-time"`
	TimeRange   int        `mapstructure:"time-range"`
	WaitTime    int        `mapstructure:"wait-time"`
	SkipTime    int        `mapstructure:"skip-time"`
	Logger      *LogConfig `mapstructure:"logger"`
}

//LogConfig system log configuration
type LogConfig struct {
	Output       string `mapstructure:"output"`
	FilePath     string `mapstructure:"file-path"`
	RotationTime int64  `mapstructure:"rotation-time"`
	MaxAge       int64  `mapstructure:"max-age"`
	Level        string `mapstructure:"level"`
}
