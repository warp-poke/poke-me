package core

import "github.com/google/go-github/github"

// WebHookPayload Github payload
type WebHookPayload struct {
	After      *string                `json:"after,omitempty"`
	Before     *string                `json:"before,omitempty"`
	Commits    []github.WebHookCommit `json:"commits,omitempty"`
	Compare    *string                `json:"compare,omitempty"`
	Created    *bool                  `json:"created,omitempty"`
	Deleted    *bool                  `json:"deleted,omitempty"`
	Forced     *bool                  `json:"forced,omitempty"`
	HeadCommit *github.WebHookCommit  `json:"head_commit,omitempty"`
	Pusher     *github.User           `json:"pusher,omitempty"`
	Ref        *string                `json:"ref,omitempty"`
	//	Repo       *Repository     `json:"repository,omitempty"`
	Sender *github.User `json:"sender,omitempty"`
}
