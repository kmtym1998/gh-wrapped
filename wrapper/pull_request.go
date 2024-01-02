package wrapper

import (
	"fmt"
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
	// 作成 ~ マージまでが最も短かった PR
	ShortestPullRequest SimplePullRequest
	// 作成 ~ マージまでが最も長かった PR
	LongestPullRequest SimplePullRequest
	// 作成 ~ マージまでの平均時間
	AverageDuration time.Duration
	// コメントが最も多かった PR
	MostCommentedPullRequest SimplePullRequest
	// コミットが最も多かった PR
	MostCommitsPullRequest SimplePullRequest
	// リポジトリごとに PR を出した数
	SubmissionRanking []SubmissionRankingItem
	// 一番レビュー回数が多かったユーザー
	MostReviewedBy string
}

type SubmissionRankingItem struct {
	PullRequest SimplePullRequest
	Count       int
}

type SimplePullRequest struct {
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
	}

	return &result, nil
}
