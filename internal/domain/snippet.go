package domain

import "errors"

var ErrSnippetIDExists = errors.New("a snippet ID already exists")

type Snippet struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Language  string   `json:"language"`
	Code      string   `json:"code"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"createdAt"`
	Favorite  bool     `json:"favorite"`
}

type CreateSnippetInput struct {
	Title    string   `json:"title"`
	Language string   `json:"language"`
	Code     string   `json:"code"`
	Tags     []string `json:"tags"`
}
