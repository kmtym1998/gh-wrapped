package wrapper

import (
	"fmt"
	"sort"
	"time"

	"github.com/kmtym1998/gh-wrapped/config"
	"github.com/kmtym1998/gh-wrapped/repository"
	"github.com/samber/lo"
)

type WrapResultPullRequest struct {
	Login string
	// 当年のすべての PR の数
	TotalCount int
	// 当年に作成され、当年にマージされた PR の数
	MergedCount int
	// 当年に作成され、当年にマージされなかった、OPEN でない PR の数
	ClosedCount int
	// 作成 ~ マージまでが最も短かった PR (上位 3 つ)
	ShortestPullRequests []PullRequestDurationItem
	// 作成 ~ マージまでが最も長かった PR (上位 3 つ)
	LongestPullRequests []PullRequestDurationItem
	// 作成 ~ マージまでの平均時間
	AverageDuration time.Duration
	// コメントが最も多くつけられた PR
	MostCommentedPullRequest SimplePullRequest
	// コミットが最も多かった PR
	MostCommitsPullRequest SimplePullRequest
	// リポジトリごとに PR を出した数
	SubmissionRanking []PullRequestSubmissionRankingItem
	// 一番レビュー回数が多かったユーザー
	MostReviewedBy string
}

type PullRequestDurationItem struct {
	PullRequest SimplePullRequest
	Duration    time.Duration
}

type PullRequestSubmissionRankingItem struct {
	PullRequest SimplePullRequest
	Count       int
}

type SimplePullRequest struct {
	Title  string
	Owner  string
	Repo   string
	Number int
	URL    string
}

func WrapPullRequest(repo repository.GitHubRepository, cfg *config.Config) (*WrapResultPullRequest, error) {
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

	result := WrapResultPullRequest{
		Login:      user.Login,
		TotalCount: len(pullRequests),
		MergedCount: lo.CountBy(pullRequests, func(pr *repository.PullRequest) bool {
			if pr.CreatedAt.Year() != cfg.Year() {
				return false
			}

			if pr.MergedAt.Valid && pr.MergedAt.Time.Year() != cfg.Year() {
				return false
			}

			return pr.State == repository.PullRequestStateMerged
		}),
		ClosedCount: lo.CountBy(pullRequests, func(pr *repository.PullRequest) bool {
			if pr.CreatedAt.Year() != cfg.Year() {
				return false
			}

			if pr.ClosedAt.Valid && pr.ClosedAt.Time.Year() != cfg.Year() {
				return false
			}

			return pr.State == repository.PullRequestStateClosed
		}),
		ShortestPullRequests: PickTopNPullRequestsDurationItem(
			pullRequests,
			func(pr1, pr2 *repository.PullRequest) bool {
				if pr1.State != repository.PullRequestStateMerged {
					return false
				}

				if !pr1.MergedAt.Valid {
					return false
				}

				if pr2.State != repository.PullRequestStateMerged {
					return false
				}

				if !pr2.MergedAt.Valid {
					return false
				}

				sub1 := pr1.MergedAt.Time.Sub(pr1.CreatedAt)
				sub2 := pr2.MergedAt.Time.Sub(pr2.CreatedAt)

				return sub1 < sub2
			},
			3,
		),
		LongestPullRequests: func() []PullRequestDurationItem {
			longestPR := lo.MaxBy(
				pullRequests,
				func(
					pr *repository.PullRequest,
					currentMax *repository.PullRequest,
				) bool {
					if pr.State != repository.PullRequestStateMerged {
						return false
					}

					if !pr.MergedAt.Valid {
						return false
					}

					newSub := pr.MergedAt.Time.Sub(pr.CreatedAt)
					currentMaxSub := currentMax.MergedAt.Time.Sub(currentMax.CreatedAt)

					return newSub > currentMaxSub
				},
			)

			return []PullRequestDurationItem{{
				PullRequest: SimplePullRequest{
					Title:  longestPR.Title,
					Owner:  longestPR.RepositoryOwner,
					Repo:   longestPR.RepositoryName,
					Number: longestPR.Number,
					URL:    longestPR.URL,
				},
				Duration: longestPR.MergedAt.Time.Sub(longestPR.CreatedAt),
			}}
		}(),
	}

	return &result, nil
}

func PickTopNPullRequestsDurationItem(
	list []*repository.PullRequest,
	compareFunc func(a, b *repository.PullRequest) bool,
	n int,
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
