package main

import "strings"

type HookDataProject struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebUrl            string `json:"web_url"`
	AvatarUrl         string `json:"avatar_url"`
	GitSSHUrl         string `json:"git_ssh_url"`
	GitHTTPUrl        string `json:"git_http_url"`
	Namespace         string `json:"namespace"`
	VisibilityLevel   int    `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
	Homepage          string `json:"homepage"`
	Url               string `json:"url"`
	SSHUrl            string `json:"ssh_url"`
	HTTPUrl           string `json:"http_url"`
}

type HookDataRepository struct {
	Name            string `json:"name"`
	Url             string `json:"url"`
	Description     string `json:"description"`
	Homepage        string `json:"homepage"`
	GitHTTPUrl      string `json:"git_http_url"`
	GitSSHUrl       string `json:"git_ssh_url"`
	VisibilityLevel int    `json:"visibility_level"`
}

type HookDataAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type HookDataCommit struct {
	Id        string         `json:"id"`
	Message   string         `json:"message"`
	Title     string         `json:"title"`
	Timestamp string         `json:"timestamp"`
	Url       string         `json:"url"`
	Author    HookDataAuthor `json:"author"`
	Added     []string       `json:"added"`
	Modified  []string       `json:"modified"`
	Removed   []string       `json:"removed"`
}

type HookData struct {
	ObjectKind        string             `json:"object_kind"`
	Before            string             `json:"before"`
	After             string             `json:"after"`
	Ref               string             `json:"ref"`
	CheckoutSha       string             `json:"checkout_sha"`
	UserId            int                `json:"user_id"`
	UserName          string             `json:"user_name"`
	UserUsername      string             `json:"user_username"`
	UserEmail         string             `json:"user_email"`
	UserAvatar        string             `json:"user_avatar"`
	ProjectId         int                `json:"project_id"`
	TotalCommitsCount int                `json:"total_commits_count"`
	Project           HookDataProject    `json:"project"`
	Repository        HookDataRepository `json:"repository"`
	Commits           []HookDataCommit   `json:"commits"`
}

func (t *HookData) GetTag() string {
	parts := strings.Split(t.Ref, "/")
	return parts[2]
}

func (t *HookData) HasFileChanged(fileName string) bool {
	for _, c := range t.Commits {
		if c.HasFileChanged(fileName) {
			return true
		}
	}
	return false
}

func (t *HookDataCommit) HasFileChanged(fileName string) bool {
	for _, p := range t.Added {
		if fileName == p {
			return true
		}
	}
	for _, p := range t.Modified {
		if fileName == p {
			return true
		}
	}
	return false
}
