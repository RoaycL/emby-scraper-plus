package model

import "time"

// ActorSearchResult represents a cast/actor search match.
type ActorSearchResult struct {
	ID       string   `json:"id"`
	Provider string   `json:"provider"`
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases"`
	Images   []string `json:"images"`
}

// ActorInfo represents detailed actor metadata.
type ActorInfo struct {
	ActorSearchResult
	Summary     string    `json:"summary,omitempty"`
	BirthDate   time.Time `json:"birth_date,omitempty"`
	BloodType   string    `json:"blood_type,omitempty"`
	Height      int       `json:"height,omitempty"`
	Nationality string    `json:"nationality,omitempty"`
	Hobby       string    `json:"hobby,omitempty"`
}
