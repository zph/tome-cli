package cmd

import (
	"fmt"
	"io"
	"log"
	"net/url"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type customWriter struct {
	io.Writer
}

func (cw customWriter) Close() error {
	return nil
}
func (cw customWriter) Sync() error {
	return nil
}

func createLogger(name string, output io.Writer) *zap.SugaredLogger {
	// Custom writer technique found here:
	// - https://github.com/uber-go/zap/issues/979
	// Allows for e2e testing of cobra application
	const customWriterKey = "cobra-writer"
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}
	config.EncoderConfig.FunctionKey = "function"

	err := zap.RegisterSink(customWriterKey, func(u *url.URL) (zap.Sink, error) {
		return customWriter{output}, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// build a valid custom path
	customPath := fmt.Sprintf("%s:io", customWriterKey)
	config.OutputPaths = []string{customPath}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger.Sugar().Named(name)
}
