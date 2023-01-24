package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus/hooks/writer"
)

func parseLogLevel(levelName string) log.Level {
	levelMap := map[string]log.Level{
		"debug": log.DebugLevel,
		"info":  log.InfoLevel,
		"warn":  log.WarnLevel,
		"error": log.ErrorLevel,
		"fatal": log.FatalLevel,
	}
	level := levelMap[levelName]
	if level == 0 {
		level = log.InfoLevel
	}
	return level
}

/*
 * Configuration must be initialized before this function called.
 */
func InitLog() {
	var cfg struct {
		Level string `json:"level" yaml:"level"`
	}
	if err := Unmarshal("log", &cfg); err != nil {
		panic(err)
	}

	log.SetLevel(parseLogLevel(cfg.Level))
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
	log.AddHook(&writer.Hook{     // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
}
