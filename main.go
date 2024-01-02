package main

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/kmtym1998/gh-wrapped/config"
	"github.com/kmtym1998/gh-wrapped/repository"
	"github.com/kmtym1998/gh-wrapped/wrapper"
	"github.com/kr/pretty"
	"github.com/m-mizutani/clog"
	"github.com/samber/lo"
)

func main() {
	cfg := config.Parse()

	// TODO: kmtym1998/handyman への以降
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
				padding := lo.Ternary(
					len(funcName) < maxLenOfFuncName,
					strings.Repeat(" ", maxLenOfFuncName-len(funcName)),
					"",
				)

				stacks = append(
					stacks,
					funcName+padding+"  "+fileAndLines[i],
				)
			}

			fmt.Print(strings.Join(stacks, "\n"))
		}
	}()

	setupLogger(cfg)

	// list := []map[string]int{
	// 	{"id": 5, "val": 3},
	// 	{"id": 3, "val": 100},
	// 	{"id": 2, "val": 2},
	// 	{"id": 4, "val": 4},
	// 	{"id": 6},
	// 	{"id": 1, "val": 1},
	// 	{"id": 7, "val": -1},
	// }

	// sort.Slice(list, func(i, j int) bool {
	// 	slog.Debug("sort", "i", list[i], "j", list[j])

	// 	if _, ok := list[i]["val"]; !ok {
	// 		return false
	// 	}

	// 	if _, ok := list[j]["val"]; !ok {
	// 		return true
	// 	}

	// 	return list[i]["val"] < list[j]["val"]
	// })

	// slog.Debug("sorted", "list", list)

	// return

	repo, err := repository.NewGitHub()
	if err != nil {
		fatal("failed to create GitHub client: %v", err)
	}

	pr, err := wrapper.WrapPullRequest(repo, cfg)
	if err != nil {
		fatal("failed to wrap pull requests: %v", err)
	}
	pretty.Println(pr)
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
	slog.Error(fmt.Sprintf(msg, args...))
	panic(msg)
}
