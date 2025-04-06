package plex

type PlexResponse struct {
	MediaContainer MediaContainer `json:"MediaContainer"`
}

type MediaContainer struct {
	Directory []Library `json:"Directory"`
}

type Library struct {
	Title string `json:"title"`
}
