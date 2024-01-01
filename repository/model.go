package repository

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type PullRequest struct {
	ID              string
	Number          int
	Title           string
	RepositoryOwner string
	RepositoryName  string
	CreatedAt       time.Time
	MergedAt        null.Time
	State           PullRequestState
	CommitsCount    int
	CommentsCount   int
	ReviewedBy      []string
	Reviews         []PullRequestReview
	URL             string
}

type PullRequestReview struct {
	ID       string
	Author   string
	State    string
	Comments []PullRequestComment
}

type PullRequestComment struct {
	ID     string
	Author string
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
