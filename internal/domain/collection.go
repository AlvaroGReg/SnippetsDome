package domain

// Collection groups snippets inside the SQLite database.
type Collection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}
