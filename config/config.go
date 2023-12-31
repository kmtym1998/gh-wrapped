package config

import (
	"flag"
	"os"
	"strings"

	"github.com/samber/lo"
)

type Config struct {
	DebugMode      bool
	IncludePrivate bool
}

func Parse() (*Config, error) {
	debugMode := strings.ToUpper(os.Getenv("DEBUG")) == "TRUE"

	flag.Parse()

	includePrivate := flag.Bool("include-private", true, "Whether to include private repositories in the aggregation")
	if includePrivate == nil {
		includePrivate = lo.ToPtr(true)
	}

	return &Config{
		DebugMode:      debugMode,
		IncludePrivate: *includePrivate,
	}, nil
}
