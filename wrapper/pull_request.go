package wrapper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kmtym1998/gh-wrapped/repository"
	"github.com/samber/lo"
)

type WrapResultPullRequest struct {
	Login string
	// 2023 年のすべての PR の数
	TotalCount int
	// 2023 年に作成され、2023 年にマージされた PR の数
	MergedCount int
	// 2023 年に作成され、2023 年にマージされなかった、OPEN でない PR の数
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
	SubmissionRanking []RankingtListItem[SimplePullRequest]
	// 一番レビュー回数が多かったユーザー
	MostReviewedBy string
}

type RankingtListItem[T any] struct {
	Item  T
	Value int
}

type SimplePullRequest struct {
	Owner  string
	Repo   string
	Number int
	URL    string
}

func WrapPullRequest(repo repository.GitHubRepository) (*WrapResultPullRequest, error) {
	from := lo.Must(time.Parse(time.RFC3339, "2023-01-01T00:00:00Z"))
	to := lo.Must(time.Parse(time.RFC3339, "2023-12-31T23:59:59Z"))
	pullRequests, err := repo.ListPullRequests(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	b, _ := json.Marshal(pullRequests)
	fmt.Println(string(b))

	return &WrapResultPullRequest{}, nil
}
