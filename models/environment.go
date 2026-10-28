package models

// Environment identifies the SDK deployment. The pinned upstream exposes only production.
type Environment struct {
	Name string `json:"name"`
}
