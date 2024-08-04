package cmd

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ValidArgsFunctionForScripts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	config := NewConfig()
	rootDir := config.RootDir()
	ignorePatterns := config.IgnorePatterns()

	if debug {
		cobra.CompDebugln(fmt.Sprintf(`completion: args=%+v, toComplete=%s`, args, toComplete), true)
	}
	// If we have an executable file in the path, we're working on completions for that script itself via --completion
	var argsAccumulator []string
	// Iteration must exit on first matching executable file or it breaks invariants of code
	for _, arg := range args {
		// __complete is passed as an internal directive
		if arg == "__complete" {
			continue
		}
		argsAccumulator = append(argsAccumulator, arg)
		joint := path.Join(append([]string{rootDir}, argsAccumulator...)...)
		if ignorePatterns.MatchesPath(joint) {
			continue
		}
		f, err := os.Stat(joint)
		if debug {
			cobra.CompDebugln(fmt.Sprintf(`completion: joint=%s`, joint), true)
			cobra.CompDebugln(fmt.Sprintf(`completion: executable=%s`, f.Mode()), true)
		}
		if err == nil && isExecutableByOwner(f.Mode()) {
			s := NewScript(joint, rootDir)
			// We have an executable file in the path
			// Handle completion for the script itself via --completion

			/*
				Check if the script contains word --completion
			*/
			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: hasCompletions=%t`, s.HasCompletions()), true)
			}
			if !s.HasCompletions() {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			/*
				Extract the completion values from the script
			*/
			// Execute the joint path as a shell script
			completionFlag := []string{"--completion"}
			output, err := exec.Command(joint, completionFlag...).Output()
			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: output=%s`, output), true)
			}
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			// Split the output into lines
			lines := strings.Split(string(output), "\n")

			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: lines=%s`, lines), true)
			}
			// Remove empty lines
			var completions []string
			for _, line := range lines {
				if debug {
					cobra.CompDebugln(fmt.Sprintf(`completion: line=%+v`, line), true)
				}
				if line != "" {
					completions = append(completions, line)
				}
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// Otherwise we're completing the path to the script
	fullPathSegments := append([]string{rootDir}, args...)
	fullPath := path.Join(fullPathSegments...)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	if len(entries) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var toCompleteEntries []string
	if toComplete == "" {
		for _, entry := range entries {
			toCompleteEntries = append(toCompleteEntries, entry.Name())
		}
	} else {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), toComplete) {
				toCompleteEntries = append(toCompleteEntries, entry.Name())
			}
		}
	}

	var executableOrDirectories []string
	for _, entry := range toCompleteEntries {
		if strings.HasSuffix(entry, "/") {
			executableOrDirectories = append(executableOrDirectories, entry)
		} else {
			fullPathWithEntry := path.Join(fullPath, entry)
			s := NewScript(fullPathWithEntry, rootDir)
			if s.IsDir() {
				executableOrDirectories = append(executableOrDirectories, entry+"\tdirectory")
			} else if s.IsExecutable() {
				if debug {
					cobra.CompDebugln(fmt.Sprintf(`completion: fullPath=%s, entry=%s, %+v`, fullPath, entry, s), true)
				}
				executableOrDirectories = append(executableOrDirectories, entry+"\t"+s.Usage())
			}
		}
	}
	return executableOrDirectories, cobra.ShellCompDirectiveNoFileComp
}

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
