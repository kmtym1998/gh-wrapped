package repository

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

type GitHubClient struct {
	restClient    *api.RESTClient
	graphQLClient *api.GraphQLClient
	host          string

	// cache
	authenticatedUser *PublicUser
}

type GitHubRepository interface {
	ListOrganizations() ([]*Organization, error)
	ListPullRequests(from, to time.Time) ([]*PullRequest, error)
	GetMe() (*PublicUser, error)
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

func (r *GitHubClient) ListPullRequests(from, to time.Time) ([]*PullRequest, error) {
	var nextCursor string
	var response WrapPullRequestsResponse
	var pullRequests []*PullRequest
	// TODO: プログレスの表示
	for {
		variables := map[string]interface{}{
			"from": from,
			"to":   to,
		}

		if nextCursor != "" {
			variables["prAfterCursor"] = nextCursor
		}

		slog.Debug(
			"getting pull requests...",
			"variables", variables,
		)

		if err := r.graphQLClient.Do(
			wrapPullRequestQuery,
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
			// TODO: review と comment のページング

			pullRequests = append(pullRequests, &PullRequest{
				ID:              node.PullRequest.ID,
				Number:          node.PullRequest.Number,
				Title:           node.PullRequest.Title,
				RepositoryOwner: node.PullRequest.Repository.Owner.Login,
				RepositoryName:  node.PullRequest.Repository.Name,
				CreatedAt:       node.PullRequest.CreatedAt,
				MergedAt:        node.PullRequest.MergedAt,
				State:           FromString(node.PullRequest.State),
				CommitsCount:    node.PullRequest.Commits.TotalCount,
				CommentsCount:   node.PullRequest.Comments.TotalCount,
				// NOTE: struct の定義がめんどくさくて lo.Map を使ってない
				Reviews: func() []PullRequestReview {
					var reviews []PullRequestReview
					for _, review := range node.PullRequest.Reviews.Nodes {
						var comments []PullRequestComment
						for _, comment := range review.Comments.Nodes {
							comments = append(comments, PullRequestComment{
								ID:     comment.ID,
								Author: comment.Author.Login,
								ReplyTo: func() string {
									// NOTE: なんかこれだと動かなかった。nil ぽが起きる
									// lo.Ternary(
									// 	comment.ReplyTo != nil,
									// 	comment.ReplyTo.ID,
									// 	"",
									// ),
									if comment.ReplyTo != nil {
										return comment.ReplyTo.ID
									}

									return ""
								}(),
							})
						}

						reviews = append(reviews, PullRequestReview{
							ID:       review.ID,
							Author:   review.Author.Login,
							State:    review.State,
							Comments: comments,
						})
					}

					return reviews
				}(),
				URL: "https://" + r.host + "/" + node.PullRequest.Repository.Owner.Login + "/" + node.PullRequest.Repository.Name + "/pull/" + strconv.Itoa(node.PullRequest.Number),
			})
		}

		if !response.Viewer.ContributionsCollection.PullRequestContributions.PageInfo.HasNextPage {
			break
		}

		nextCursor = response.Viewer.ContributionsCollection.PullRequestContributions.PageInfo.EndCursor

		// レートリミット対策のために sleep
		time.Sleep(1 * time.Second)
	}

	return pullRequests, nil
}
