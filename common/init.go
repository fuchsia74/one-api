package common

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
)

var (
	Port         = flag.Int("port", 3000, "the listening port")
	PrintVersion = flag.Bool("version", false, "print version and exit")
	PrintHelp    = flag.Bool("help", false, "print help and exit")
	LogDir       = flag.String("log-dir", "./logs", "specify the log directory")
)

// func printHelp() {
// 	fmt.Println("One API " + Version + " - All in one API service for OpenAI API.")
// 	fmt.Println("Copyright (C) 2025 JustSong. All rights reserved.")
// 	fmt.Println("GitHub: https://github.com/Laisky/one-api")
// 	fmt.Println("Usage: one-api [--port <port>] [--log-dir <log directory>] [--version] [--help]")
// }

func Init() {
	flag.Parse()

	// if *PrintVersion {
	// 	fmt.Println(Version)
	// 	os.Exit(0)
	// }

	// if *PrintHelp {
	// 	printHelp()
	// 	os.Exit(0)
	// }

	if config.SessionSecretEnvValue != "" {
		if config.SessionSecretEnvValue == "random_string" {
			logger.Logger.Error("SESSION_SECRET is set to an example value, please change it to a random string.")
		} else {
			config.SessionSecret = config.SessionSecretEnvValue
		}
	}
	SQLitePath = config.SQLitePath
	if *LogDir != "" {
		expanded := expandLogDirPath(*LogDir)
		lg := logger.Logger.With(zap.String("log_dir", expanded))
		lg.Debug("starting to set log dir")

		var err error
		expanded, err = filepath.Abs(expanded)
		if err != nil {
			lg.Fatal("failed to get absolute log dir", zap.Error(err))
		}

		if err = os.MkdirAll(expanded, 0o777); err != nil {
			lg.Fatal("failed to create log dir", zap.Error(err))
		}

		lg.Info("set log dir", zap.String("log_dir", expanded))
		logger.LogDir = expanded
		*LogDir = expanded
	}
}
