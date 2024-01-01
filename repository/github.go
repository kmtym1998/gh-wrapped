package repository

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/volatiletech/null/v8"
)

type GitHubClient struct {
	restClient    *api.RESTClient
	graphQLClient *api.GraphQLClient
	host          string
}

type GitHubRepository interface {
	ListOrganizations() ([]*Organization, error)
	ListPullRequests(from, to time.Time) ([]*PullRequest, error)
}

func NewGitHub() (*GitHubClient, error) {
	host, _ := auth.DefaultHost()

	rest, err := api.NewRESTClient(
		api.ClientOptions{
			Timeout: 10 * time.Second,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}

	graphql, err := api.NewGraphQLClient(
		api.ClientOptions{
			Timeout: 10 * time.Second,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	return &GitHubClient{
		restClient:    rest,
		graphQLClient: graphql,
		host:          host,
	}, nil
}

type Organization struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"nodeId"`
	URL              string `json:"url"`
	ReposURL         string `json:"reposUrl"`
	EventsURL        string `json:"eventsUrl"`
	HooksURL         string `json:"hooksUrl"`
	IssuesURL        string `json:"issuesUrl"`
	MembersURL       string `json:"membersUrl"`
	PublicMembersURL string `json:"publicMembersUrl"`
	AvatarURL        string `json:"avatarUrl"`
	Description      string `json:"description"`
}

// https://docs.github.com/ja/rest/orgs/orgs?apiVersion=2022-11-28#list-organizations
func (r *GitHubClient) ListOrganizations() ([]*Organization, error) {
	var response []*Organization
	err := r.restClient.Get("user/orgs", &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type PullRequest struct {
	ID              string           `json:"id"`
	Number          int              `json:"number"`
	Title           string           `json:"title"`
	RepositoryOwner string           `json:"repositoryOwner"`
	RepositoryName  string           `json:"repositoryName"`
	CreatedAt       time.Time        `json:"createdAt"`
	MergedAt        null.Time        `json:"mergedAt"`
	State           PullRequestState `json:"state"`
	CommitsCount    int              `json:"commitsCount"`
	CommentsCount   int              `json:"commentsCount"`
	URL             string           `json:"url"`
}

type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "OPEN"
	PullRequestStateClosed PullRequestState = "CLOSED"
	PullRequestStateMerged PullRequestState = "MERGED"
)

func (s PullRequestState) String() string {
	return string(s)
}

func FromString(s string) PullRequestState {
	switch s {
	case "OPEN":
		return PullRequestStateOpen
	case "CLOSED":
		return PullRequestStateClosed
	case "MERGED":
		return PullRequestStateMerged
	default:
		panic("invalid PullRequestState")
	}
}

func (r *GitHubClient) ListPullRequests(from, to time.Time) ([]*PullRequest, error) {
	const query = `
query Wrapped($from: DateTime, $to: DateTime, $afterCursor: String) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      pullRequestContributions(first: 100, after: $afterCursor) {
        totalCount
        pageInfo {
          startCursor
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
          }
        }
      }
    }
  }
}`

	var cursor string
	var response struct {
		Viewer struct {
			ContributionsCollection struct {
				PullRequestContributions struct {
					TotalCount int `json:"totalCount"`
					PageInfo   struct {
						StartCursor string `json:"startCursor"`
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
					Nodes []struct {
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
						} `json:"pullRequest"`
					} `json:"nodes"`
				} `json:"pullRequestContributions"`
			} `json:"contributionsCollection"`
		} `json:"viewer"`
	}
	var pullRequests []*PullRequest
	// TODO: プログレスの表示
	for {
		variables := map[string]interface{}{
			"from": from,
			"to":   to,
		}

		if cursor != "" {
			variables["afterCursor"] = cursor
		}

		slog.Debug(
			"getting pull requests...",
			"variables", variables,
		)

		if err := r.graphQLClient.Do(
			query,
			variables,
			&response,
		); err != nil {
			return nil, err
		}

		slog.Debug("request done!")

		if response.Viewer.ContributionsCollection.PullRequestContributions.TotalCount == 0 {
			break
		}

		for _, node := range response.Viewer.ContributionsCollection.PullRequestContributions.Nodes {
			slog.Debug("debug", "title", node.PullRequest.Title, "createdAt", node.PullRequest.CreatedAt, "mergedAt", node.PullRequest.MergedAt)

			pullRequests = append(pullRequests, &PullRequest{
				ID:              node.PullRequest.ID,
				Number:          node.PullRequest.Number,
				Title:           node.PullRequest.Title,
				RepositoryOwner: node.PullRequest.Repository.Owner.Login,
				RepositoryName:  node.PullRequest.Repository.Name,
				CreatedAt:       node.PullRequest.CreatedAt,
				MergedAt:        node.PullRequest.MergedAt,
				CommitsCount:    node.PullRequest.Commits.TotalCount,
				CommentsCount:   node.PullRequest.Comments.TotalCount,
				State:           FromString(node.PullRequest.State),
				URL:             "https://" + r.host + "/" + node.PullRequest.Repository.Owner.Login + "/" + node.PullRequest.Repository.Name + "/pull/" + strconv.Itoa(node.PullRequest.Number),
			})
		}

		if response.Viewer.ContributionsCollection.PullRequestContributions.PageInfo.HasNextPage {
			cursor = response.Viewer.ContributionsCollection.PullRequestContributions.PageInfo.EndCursor

			// レートリミット対策のために sleep
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	return pullRequests, nil
}
