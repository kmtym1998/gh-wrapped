package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type Config struct {
	DebugMode      bool
	IncludePrivate bool
	year           int
}

func Parse() *Config {
	debugMode := strings.ToUpper(os.Getenv("DEBUG")) == "TRUE"

	flag.Parse()

	includePrivate := flag.Bool("include-private", true, "Whether to include private repositories in the aggregation")
	if includePrivate == nil {
		includePrivate = lo.ToPtr(true)
	}

	return &Config{
		DebugMode:      debugMode,
		IncludePrivate: *includePrivate,
	}
}

func (c *Config) Year() int {
	if c.year == 0 {
		return 2021
	}
	return c.year
}

func (c *Config) YearString() string {
	return strconv.Itoa(c.Year())
}
