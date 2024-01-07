package wrapper

import (
	"fmt"
	"sort"
	"time"

	"github.com/kmtym1998/gh-wrapped/config"
	"github.com/kmtym1998/gh-wrapped/repository"
	"github.com/montanaflynn/stats"
	"github.com/samber/lo"
)

type WrappedResultPullRequest struct {
	Login string
	// 当年のすべての PR の数
	TotalCount int
	// 当年に作成され、当年にマージされた PR の数
	MergedCount int
	// 当年に作成され、当年にマージされなかった、OPEN でない PR の数
	ClosedCount int
	// 作成 ~ マージまでが最も短かった PR (上位 3 つ)
	ShortLivePullRequests []PullRequestDurationItem
	// 作成 ~ マージまでが最も長かった PR (上位 3 つ)
	LongLiveRequests []PullRequestDurationItem
	// 作成 ~ マージまでの平均時間
	DurationStats PullRequestDuration
	// コメントが最も多くつけられた PR
	MostCommentedPullRequests []PullRequestRankingItem
	// コミットが最も多かった PR
	MostCommittedPullRequests []PullRequestRankingItem
	// リポジトリごとに PR を出した数
	SubmissionRanking []PullRequestRankingItem
	// 一番レビュー回数が多かったユーザー
	MostReviewedBy string
}

type PullRequestDurationItem struct {
	PullRequest SimplePullRequest
	Duration    time.Duration
}

type PullRequestRankingItem struct {
	PullRequest SimplePullRequest
	Count       int
}

type PullRequestDuration struct {
	Average      time.Duration
	Min          time.Duration
	Percentile50 time.Duration
	Percentile90 time.Duration
	Percentile99 time.Duration
	Max          time.Duration
}

type SimplePullRequest struct {
	Title  string
	Owner  string
	Repo   string
	Number int
	URL    string
}

func WrapPullRequest(repo repository.GitHubRepository, cfg *config.Config) (*WrappedResultPullRequest, error) {
	user, err := repo.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	pullRequests, err := repo.ListPullRequests(
		lo.Must(time.Parse(time.RFC3339, cfg.YearString()+"-01-01T00:00:00Z")),
		lo.Must(time.Parse(time.RFC3339, cfg.YearString()+"-12-31T23:59:59Z")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	result := WrappedResultPullRequest{
		Login:      user.Login,
		TotalCount: len(pullRequests),
		MergedCount: countPullRequestsMergedThisYear(
			pullRequests,
			cfg.Year(),
		),
		ClosedCount: lo.CountBy(pullRequests, func(pr *repository.PullRequest) bool {
			if pr.CreatedAt.Year() != cfg.Year() {
				return false
			}

			if pr.ClosedAt.Valid && pr.ClosedAt.Time.Year() != cfg.Year() {
				return false
			}

			return pr.State == repository.PullRequestStateClosed
		}),
		ShortLivePullRequests: pickTopNPullRequestsDurationItemAsc(
			pullRequests,
			3,
			func(pr1, pr2 *repository.PullRequest) bool {
				if pr1.State != repository.PullRequestStateMerged {
					return false
				}

				if !pr1.MergedAt.Valid {
					return false
				}

				if pr2.State != repository.PullRequestStateMerged {
					return true
				}

				if !pr2.MergedAt.Valid {
					return true
				}

				sub1 := pr1.MergedAt.Time.Sub(pr1.CreatedAt)
				sub2 := pr2.MergedAt.Time.Sub(pr2.CreatedAt)

				return sub1 < sub2
			},
		),
		LongLiveRequests: pickTopNPullRequestsDurationItemAsc(
			pullRequests,
			3,
			func(pr1, pr2 *repository.PullRequest) bool {
				if pr1.State != repository.PullRequestStateMerged {
					return false
				}

				if !pr1.MergedAt.Valid {
					return false
				}

				if pr2.State != repository.PullRequestStateMerged {
					return true
				}

				if !pr2.MergedAt.Valid {
					return true
				}

				sub1 := pr1.MergedAt.Time.Sub(pr1.CreatedAt)
				sub2 := pr2.MergedAt.Time.Sub(pr2.CreatedAt)

				return sub1 > sub2
			},
		),
		DurationStats: func() PullRequestDuration {
			// average
			sum := lo.SumBy(pullRequests, func(pr *repository.PullRequest) time.Duration {
				if pr.State != repository.PullRequestStateMerged {
					return 0
				}

				if !pr.MergedAt.Valid {
					return 0
				}

				return pr.MergedAt.Time.Sub(pr.CreatedAt)
			})
			avg := sum / time.Duration(countPullRequestsMergedThisYear(pullRequests, cfg.Year()))

			prLifetimes := lo.FilterMap(
				pullRequests,
				func(pr *repository.PullRequest, _ int) (float64, bool) {
					if pr.State != repository.PullRequestStateMerged {
						return 0, false
					}

					if !pr.MergedAt.Valid {
						return 0, false
					}

					return float64(pr.MergedAt.Time.Sub(pr.CreatedAt)), true
				},
			)

			return PullRequestDuration{
				Average: avg,
				Min:     time.Duration(lo.Min(prLifetimes)),
				Percentile50: time.Duration(
					lo.Must(stats.Percentile(prLifetimes, 50)),
				) * time.Nanosecond,
				Percentile90: time.Duration(
					lo.Must(stats.Percentile(prLifetimes, 90)),
				) * time.Nanosecond,
				Percentile99: time.Duration(
					lo.Must(stats.Percentile(prLifetimes, 99)),
				) * time.Nanosecond,
				Max: time.Duration(lo.Max(prLifetimes)),
			}
		}(),
		MostCommentedPullRequests: pickTopNPullRequestRankingItemDesc(
			pullRequests,
			3,
			func(pr *repository.PullRequest) int {
				return pr.CommentsCount
			},
		),
		MostCommittedPullRequests: pickTopNPullRequestRankingItemDesc(
			pullRequests,
			3,
			func(pr *repository.PullRequest) int {
				return pr.CommitsCount
			},
		),
	}

	return &result, nil
}

// valueFunc で指定した値の降順で並べた上で、上位 n 件を返す
func pickTopNPullRequestRankingItemDesc(
	list []*repository.PullRequest,
	n int,
	valueFunc func(pr *repository.PullRequest) int,
) []PullRequestRankingItem {
	if n < 1 {
		panic("n must be greater than 0")
	}

	if valueFunc == nil {
		panic("compareFunc must not be nil")
	}

	var copiedList []*repository.PullRequest
	for _, pr := range list {
		copiedList = append(copiedList, pr)
	}

	sort.SliceStable(copiedList, func(i, j int) bool {
		return valueFunc(copiedList[i]) > valueFunc(copiedList[j])
	})

	var result []PullRequestRankingItem
	for i := 0; i < n; i++ {
		result = append(result, PullRequestRankingItem{
			PullRequest: SimplePullRequest{
				Title:  copiedList[i].Title,
				Owner:  copiedList[i].RepositoryOwner,
				Repo:   copiedList[i].RepositoryName,
				Number: copiedList[i].Number,
				URL:    copiedList[i].URL,
			},
			Count: valueFunc(copiedList[i]),
		})
	}

	return result
}

// compareFunc で指定した順に昇順で並べた上で、上位 n 件を返す
func pickTopNPullRequestsDurationItemAsc(
	list []*repository.PullRequest,
	n int,
	compareFunc func(a, b *repository.PullRequest) bool,
) []PullRequestDurationItem {
	if n < 1 {
		panic("n must be greater than 0")
	}

	if compareFunc == nil {
		panic("compareFunc must not be nil")
	}

	var copiedList []*repository.PullRequest
	for _, pr := range list {
		copiedList = append(copiedList, pr)
	}

	sort.SliceStable(copiedList, func(i, j int) bool {
		return compareFunc(copiedList[i], copiedList[j])
	})

	var result []PullRequestDurationItem
	for i := 0; i < n; i++ {
		result = append(result, PullRequestDurationItem{
			PullRequest: SimplePullRequest{
				Title:  copiedList[i].Title,
				Owner:  copiedList[i].RepositoryOwner,
				Repo:   copiedList[i].RepositoryName,
				Number: copiedList[i].Number,
				URL:    copiedList[i].URL,
			},
			Duration: copiedList[i].MergedAt.Time.Sub(copiedList[i].CreatedAt),
		})
	}

	return result
}

// 該当の年に作成され、マージされた PR の数を返す
func countPullRequestsMergedThisYear(pullRequests []*repository.PullRequest, year int) int {
	return lo.CountBy(pullRequests, func(pr *repository.PullRequest) bool {
		if pr.CreatedAt.Year() != year {
			return false
		}

		if pr.MergedAt.Valid && pr.MergedAt.Time.Year() != year {
			return false
		}

		return pr.State == repository.PullRequestStateMerged
	})
}
