package config

import (
	"flag"

	"github.com/samber/lo"
)

type Config struct {
	IncludePrivate bool
}

func Parse() (*Config, error) {
	flag.Parse()

	includePrivate := flag.Bool("include-private", true, "Whether to include private repositories in the aggregation")
	if includePrivate == nil {
		includePrivate = lo.ToPtr(true)
	}

	return &Config{
		IncludePrivate: *includePrivate,
	}, nil
}
