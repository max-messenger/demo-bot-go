package info

import (
	"encoding/json"
	"fmt"
	"net/http"

	"demo_bot/pkg/marshaler"
)

var ( // set by ld flags.
	CommitAuthor    string
	CommitShortSHA  string
	CommitTimestamp string
	CommitMessage   string
	CommitTag       string
	CommitRefName   string
)

type GitCommitUser struct {
	Name string `json:"name"`
}

type GitCommitMessage struct {
	Full string `json:"full"`
}

type GitCommit struct {
	User    GitCommitUser    `json:"user"`
	ID      string           `json:"id"`
	Time    string           `json:"time"`
	Message GitCommitMessage `json:"message"`
}

type Git struct {
	Commit GitCommit `json:"commit"`
	Branch string    `json:"branch"`
}

type Build struct {
	Artifact string `json:"artifact"`
	Time     string `json:"time"`
	Version  string `json:"version"`
}

type Response struct {
	Git Git `json:"git"`

	Build Build `json:"build"`
}

type Info struct {
	info json.RawMessage
}

func NewInfo(cfg Config) (*Info, error) {
	inf := Response{
		Git: Git{
			Commit: GitCommit{
				User: GitCommitUser{
					Name: CommitAuthor,
				},
				ID:   CommitShortSHA,
				Time: CommitTimestamp,
				Message: GitCommitMessage{
					Full: CommitMessage,
				},
			},
			Branch: CommitRefName,
		},
		Build: Build{
			Artifact: cfg.AppName,
			Time:     CommitTimestamp,
			Version:  cfg.BuildVersion(),
		},
	}
	rm, err := marshaler.MarshalJSON(inf)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return &Info{
		info: rm,
	}, nil
}

func (i *Info) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(i.info); err != nil {
			return
		}
	}
}
