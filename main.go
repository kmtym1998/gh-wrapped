package main

import (
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/kmtym1998/gh-wrapped/config"
	"github.com/m-mizutani/clog"
	"github.com/samber/lo"
)

func main() {
	cfg := config.Parse()
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic", "error", err)

			var funcNames, fileAndLines []string
			for depth := 0; ; depth++ {
				pc, file, line, ok := runtime.Caller(depth)
				if !ok {
					break
				}
				f := runtime.FuncForPC(pc)

				funcNames = append(funcNames, f.Name())
				fileAndLines = append(fileAndLines, fmt.Sprintf("%s:%d", file, line))
			}

			longestFuncName := lo.MaxBy(funcNames, func(a, b string) bool {
				return len(a) > len(b)
			})
			maxLenOfFuncName := len(longestFuncName)

			var stacks []string
			for i, funcName := range funcNames {
				slog.Info(strconv.Itoa(i), "funcName", funcName, "len(funcName)", len(funcName), "maxLenOfFuncName", maxLenOfFuncName)
				prefix := ""
				if len(funcName) < maxLenOfFuncName {
					prefix += funcName + strings.Repeat(" ", maxLenOfFuncName-len(funcName))
					log.Println("0 prefix", prefix)
				} else {
					prefix += funcName
					log.Println("1 prefix", prefix)
				}

				suffix := fileAndLines[i]

				stacks = append(stacks, fmt.Sprintf("%s %s", prefix, suffix))
			}

			fmt.Print(strings.Join(stacks, "\n"))
		}
	}()

	setupLogger(cfg)

	panic("test")

	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	response := struct{ Login string }{}
	err = client.Get("user", &response)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("running as %s\n", response.Login)
}

func setupLogger(cfg *config.Config) {
	logger := slog.New(
		clog.New(
			clog.WithSource(cfg.DebugMode),
			clog.WithLevel(
				lo.Ternary(
					cfg.DebugMode,
					slog.LevelDebug,
					slog.LevelInfo,
				),
			),
		),
	)

	slog.SetDefault(logger)
}

func fatal(msg string, args ...interface{}) {
	slog.Error(msg, args...)
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
