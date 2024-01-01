package repository

import (
	"time"

	"github.com/volatiletech/null/v8"
)

const wrapPullRequestQuery = `
query WrapPullRequest($from: DateTime, $to: DateTime, $prAfterCursor: String) {
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
            comments {
              totalCount
            }
            state
            createdAt
            mergedAt
            reviews(first: 50) {
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
                comments(first: 50) {
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
					PullRequest struct {
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
						Comments struct {
							TotalCount int `json:"totalCount"`
						} `json:"comments"`
						State     string    `json:"state"`
						CreatedAt time.Time `json:"createdAt"`
						MergedAt  null.Time `json:"mergedAt"`
						Reviews   struct {
							TotalCount int      `json:"totalCount"`
							PageInfo   PageInfo `json:"pageInfo"`
							Nodes      []struct {
								ID     string `json:"id"`
								State  string `json:"state"`
								Author struct {
									Login string `json:"login"`
								} `json:"author"`
								Comments struct {
									PageInfo PageInfo `json:"pageInfo"`
									Nodes    []struct {
										ID      string `json:"id"`
										ReplyTo *struct {
											ID string `json:"id"`
										} `json:"replyTo"`
										Author struct {
											Login string `json:"login"`
										} `json:"author"`
									} `json:"nodes"`
								} `json:"comments"`
							} `json:"nodes"`
						} `json:"reviews"`
					} `json:"pullRequest"`
				} `json:"nodes"`
			} `json:"pullRequestContributions"`
		} `json:"contributionsCollection"`
	} `json:"viewer"`
}
