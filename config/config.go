package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DebugMode bool
	year      int
}

func Parse() *Config {
	debugMode := strings.ToUpper(os.Getenv("DEBUG")) == "TRUE"

	flag.Parse()

	return &Config{
		DebugMode: debugMode,
		year:      2023,
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
