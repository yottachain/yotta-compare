package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ytcompare "github.com/yottachain/yotta-compare"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yotta-compare",
	Short: "compare differences of shards between SN and miner",
	Long:  `yotta-compare is an comparison service performing shards differences comparing between SN and miner.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		config := new(ytcompare.Config)
		if err := viper.Unmarshal(config); err != nil {
			panic(fmt.Sprintf("unable to decode into config struct, %v\n", err))
		}
		initLog(config)
		compare, err := ytcompare.New(context.Background(), config)
		if err != nil {
			panic(fmt.Sprintf("fatal error when starting compare service: %s\n", err))
		}
		compare.Start(context.Background())
		// config := new(ytsync.Config)
		// if err := viper.Unmarshal(config); err != nil {
		// 	panic(fmt.Sprintf("unable to decode into config struct, %v\n", err))
		// }
		// initLog(config)
		// rebuilder, err := ytsync.New(config.AnalysisDBURL, config.RebuilderDBURL, config.AuraMQ, config.Compensation, config.MiscConfig)
		// if err != nil {
		// 	panic(fmt.Sprintf("fatal error when starting rebuilder service: %s\n", err))
		// }
		// rebuilder.Start()
		// lis, err := net.Listen("tcp", config.BindAddr)
		// if err != nil {
		// 	log.Fatalf("failed to listen address %s: %s\n", config.BindAddr, err)
		// }
		// log.Infof("GRPC address: %s", config.BindAddr)
		// grpcServer := grpc.NewServer()
		// server := &ytsync.Server{Rebuilder: rebuilder}
		// pb.RegisterRebuilderServer(grpcServer, server)
		// grpcServer.Serve(lis)
		// log.Info("GRPC server started")
	},
}

func initLog(config *ytcompare.Config) {
	switch strings.ToLower(config.Logger.Output) {
	case "file":
		writer, _ := rotatelogs.New(
			config.Logger.FilePath+".%Y%m%d",
			rotatelogs.WithLinkName(config.Logger.FilePath),
			rotatelogs.WithMaxAge(time.Duration(config.Logger.MaxAge)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(config.Logger.RotationTime)*time.Hour),
		)
		log.SetOutput(writer)
	case "stdout":
		log.SetOutput(os.Stdout)
	default:
		fmt.Printf("no such option: %s, use stdout\n", config.Logger.Output)
		log.SetOutput(os.Stdout)
	}
	log.SetFormatter(&log.TextFormatter{})
	levelMap := make(map[string]log.Level)
	levelMap["panic"] = log.PanicLevel
	levelMap["fatal"] = log.FatalLevel
	levelMap["error"] = log.ErrorLevel
	levelMap["warn"] = log.WarnLevel
	levelMap["info"] = log.InfoLevel
	levelMap["debug"] = log.DebugLevel
	levelMap["trace"] = log.TraceLevel
	log.SetLevel(levelMap[strings.ToLower(config.Logger.Level)])
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/yotta-compare.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initFlag()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".yotta-rebuilder" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("yotta-compare")
		viper.SetConfigType("yaml")
	}

	// viper.AutomaticEnv() // read in environment variables that match
	// viper.SetEnvPrefix("analysis")
	// viper.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("Config file not found.")
		} else {
			// Config file was found but another error was produced
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
	}
}

var (
	//DefaultMongoDBURL default value of MongoDBURL
	DefaultMongoDBURL string = "mongodb://127.0.0.1:27017/?connect=direct"
	//DefaultDBName default value of DBName
	DefaultDBName string = "compare"
	//DefaultAllSyncURLs default value of AllSyncURLs
	DefaultAllSyncURLs []string = []string{}
	//DefaultStartTime default value of StartTime
	DefaultStartTime int = 0
	//DefaultTimeRange default value of TimeRange
	DefaultTimeRange int = 600
	//DefaultWaitTime default value of WaitTime
	DefaultWaitTime int = 60
	//DefaultSkipTime default value of SkipTime
	DefaultSkipTime int = 300

	//DefaultCOSSchema default value of COSSchema
	DefaultCOSSchema string = "https"
	//DefaultCOSDomain default value of COSDomain
	DefaultCOSDomain string = "cos.ap-beijing.myqcloud.com"
	//DefaultCOSBucketName default value of COSBucketName
	DefaultCOSBucketName string = "compare"
	//DefaultCOSSecretID default value of COSSecretID
	DefaultCOSSecretID string = ""
	//DefaultCOSSecretKey default value of COSSecretKey
	DefaultCOSSecretKey string = ""

	//DefaultLoggerOutput default value of LoggerOutput
	DefaultLoggerOutput string = "stdout"
	//DefaultLoggerFilePath default value of LoggerFilePath
	DefaultLoggerFilePath string = "./compare.log"
	//DefaultLoggerRotationTime default value of LoggerRotationTime
	DefaultLoggerRotationTime int64 = 24
	//DefaultLoggerMaxAge default value of LoggerMaxAge
	DefaultLoggerMaxAge int64 = 240
	//DefaultLoggerLevel default value of LoggerLevel
	DefaultLoggerLevel string = "Info"
)

func initFlag() {
	//config
	rootCmd.PersistentFlags().String(ytcompare.MongoDBURLField, DefaultMongoDBURL, "URL of mongoDB")
	viper.BindPFlag(ytcompare.MongoDBURLField, rootCmd.PersistentFlags().Lookup(ytcompare.MongoDBURLField))
	rootCmd.PersistentFlags().String(ytcompare.DBNameField, DefaultDBName, "name of database")
	viper.BindPFlag(ytcompare.DBNameField, rootCmd.PersistentFlags().Lookup(ytcompare.DBNameField))
	rootCmd.PersistentFlags().StringSlice(ytcompare.AllSyncURLsField, DefaultAllSyncURLs, "all URLs of sync services, in the form of --all-sync-urls \"URL1,URL2,URL3\"")
	viper.BindPFlag(ytcompare.AllSyncURLsField, rootCmd.PersistentFlags().Lookup(ytcompare.AllSyncURLsField))
	rootCmd.PersistentFlags().Int(ytcompare.StartTimeField, DefaultStartTime, "get shards from this timestamp")
	viper.BindPFlag(ytcompare.StartTimeField, rootCmd.PersistentFlags().Lookup(ytcompare.StartTimeField))
	rootCmd.PersistentFlags().Int(ytcompare.TimeRangeField, DefaultTimeRange, "time range when fetching shards for comparing")
	viper.BindPFlag(ytcompare.TimeRangeField, rootCmd.PersistentFlags().Lookup(ytcompare.TimeRangeField))
	rootCmd.PersistentFlags().Int(ytcompare.WaitTimeField, DefaultWaitTime, "wait time when no new shards can be fetched")
	viper.BindPFlag(ytcompare.WaitTimeField, rootCmd.PersistentFlags().Lookup(ytcompare.WaitTimeField))
	rootCmd.PersistentFlags().Int(ytcompare.SkipTimeField, DefaultSkipTime, "ensure not to fetching shards till the end")
	viper.BindPFlag(ytcompare.SkipTimeField, rootCmd.PersistentFlags().Lookup(ytcompare.SkipTimeField))
	//COS config
	rootCmd.PersistentFlags().String(ytcompare.COSSchemaField, DefaultCOSSchema, "schema of COS connection")
	viper.BindPFlag(ytcompare.COSSchemaField, rootCmd.PersistentFlags().Lookup(ytcompare.COSSchemaField))
	rootCmd.PersistentFlags().String(ytcompare.COSDomainField, DefaultCOSDomain, "domain name of COS")
	viper.BindPFlag(ytcompare.COSDomainField, rootCmd.PersistentFlags().Lookup(ytcompare.COSDomainField))
	rootCmd.PersistentFlags().String(ytcompare.COSBucketNameField, DefaultCOSBucketName, "bucket name of COS")
	viper.BindPFlag(ytcompare.COSBucketNameField, rootCmd.PersistentFlags().Lookup(ytcompare.COSBucketNameField))
	rootCmd.PersistentFlags().String(ytcompare.COSSecretIDField, DefaultCOSSecretID, "secret ID of COS")
	viper.BindPFlag(ytcompare.COSSecretIDField, rootCmd.PersistentFlags().Lookup(ytcompare.COSSecretIDField))
	rootCmd.PersistentFlags().String(ytcompare.COSSecretKeyField, DefaultCOSSecretKey, "secret key of COS")
	viper.BindPFlag(ytcompare.COSSecretKeyField, rootCmd.PersistentFlags().Lookup(ytcompare.COSSecretKeyField))
	//logger config
	rootCmd.PersistentFlags().String(ytcompare.LoggerOutputField, DefaultLoggerOutput, "Output type of logger(stdout or file)")
	viper.BindPFlag(ytcompare.LoggerOutputField, rootCmd.PersistentFlags().Lookup(ytcompare.LoggerOutputField))
	rootCmd.PersistentFlags().String(ytcompare.LoggerFilePathField, DefaultLoggerFilePath, "Output path of log file")
	viper.BindPFlag(ytcompare.LoggerFilePathField, rootCmd.PersistentFlags().Lookup(ytcompare.LoggerFilePathField))
	rootCmd.PersistentFlags().Int64(ytcompare.LoggerRotationTimeField, DefaultLoggerRotationTime, "Rotation time(hour) of log file")
	viper.BindPFlag(ytcompare.LoggerRotationTimeField, rootCmd.PersistentFlags().Lookup(ytcompare.LoggerRotationTimeField))
	rootCmd.PersistentFlags().Int64(ytcompare.LoggerMaxAgeField, DefaultLoggerMaxAge, "Within the time(hour) of this value each log file will be kept")
	viper.BindPFlag(ytcompare.LoggerMaxAgeField, rootCmd.PersistentFlags().Lookup(ytcompare.LoggerMaxAgeField))
	rootCmd.PersistentFlags().String(ytcompare.LoggerLevelField, DefaultLoggerLevel, "Log level(Trace, Debug, Info, Warning, Error, Fatal, Panic)")
	viper.BindPFlag(ytcompare.LoggerLevelField, rootCmd.PersistentFlags().Lookup(ytcompare.LoggerLevelField))
}
