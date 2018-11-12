package utils

// Configuration contains some sensitive passwords and usernames
type Configuration struct {
	GithubPersonalAccessToken string `json:"GithubPersonalAccessToken"`
	GithubOwner               string `json:"GithubOwner"`
	GithubRepo                string `json:"GithubRepo"`
}
