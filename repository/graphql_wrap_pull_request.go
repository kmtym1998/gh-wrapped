package repository

import (
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	reviewsLimit        = 50
	reviewCommentsLimit = 50
)

const wrapPullRequestQuery = `
query WrapPullRequest($from: DateTime, $to: DateTime, $prAfterCursor: String, $reviewsLimit: Int = 50, $reviewCommentsLimit: Int = 50) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      pullRequestContributions(first: 100, after: $prAfterCursor) {
        totalCount
        pageInfo {
          endCursor
          hasNextPage
        }
        nodes {
          pullRequest {
            id
            number
            title
            repository {
              owner {
                id
                login
              }
              name
            }
            commits {
              totalCount
            }
            state
            createdAt
            closedAt
            mergedAt
            reviews(first: $reviewsLimit) {
              totalCount
              pageInfo {
                endCursor
                hasNextPage
              }
              nodes {
                id
                state
                author {
                  login
                }
                comments(first: $reviewCommentsLimit) {
                  pageInfo {
                    endCursor
                    hasNextPage
                  }
                  nodes {
                    id
                    replyTo {
                      id
                    }
                    author {
                      login
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

type WrapPullRequestsResponse struct {
	Viewer struct {
		ContributionsCollection struct {
			PullRequestContributions struct {
				TotalCount int      `json:"totalCount"`
				PageInfo   PageInfo `json:"pageInfo"`
				Nodes      []struct {
					PullRequest PullRequestNode `json:"pullRequest"`
				} `json:"nodes"`
			} `json:"pullRequestContributions"`
		} `json:"contributionsCollection"`
	} `json:"viewer"`
}

type PullRequestNode struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Repository struct {
		Owner struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	} `json:"repository"`
	Commits struct {
		TotalCount int `json:"totalCount"`
	} `json:"commits"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	ClosedAt  null.Time `json:"closedAt"`
	MergedAt  null.Time `json:"mergedAt"`
	Reviews   struct {
		TotalCount int          `json:"totalCount"`
		PageInfo   PageInfo     `json:"pageInfo"`
		Nodes      []ReviewNode `json:"nodes"`
	} `json:"reviews"`
}

type ReviewNode struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Comments struct {
		PageInfo PageInfo            `json:"pageInfo"`
		Nodes    []ReviewCommentNode `json:"nodes"`
	} `json:"comments"`
}

type ReviewCommentNode struct {
	ID      string `json:"id"`
	ReplyTo *struct {
		ID string `json:"id"`
	} `json:"replyTo"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

func (n PullRequestNode) CommentsCount() int {
	for _, review := range n.Reviews.Nodes {
		return len(review.Comments.Nodes)
	}

	return 0
}
