package statsfm

import "github.com/Jayllyz/statsfm-card/internal/config"

// TopResponse is the stats.fm "top items" API envelope.
type TopResponse struct {
	Items []Item `json:"items"`
}

// Entity is a named, imaged stats.fm object (artist or album).
type Entity struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Track is a stats.fm track, which carries its cover art via its albums.
type Track struct {
	Name   string   `json:"name"`
	Image  string   `json:"image"`
	Albums []Entity `json:"albums"`
}

// Item is one entry in a top-items response. Exactly one of Artist,
// Track, or Album is populated depending on the requested type.
type Item struct {
	Artist   *Entity `json:"artist,omitempty"`
	Track    *Track  `json:"track,omitempty"`
	Album    *Entity `json:"album,omitempty"`
	PlayedMs *int64  `json:"playedMs,omitempty"`
	Streams  *int64  `json:"streams,omitempty"`
}

// NameAndImage returns the display name and cover image for the item,
// falling back to the track's first album image for tracks without
// their own image.
func (i Item) NameAndImage(itemType string) (name, image string) {
	switch itemType {
	case config.TypeArtists:
		if i.Artist != nil {
			return i.Artist.Name, i.Artist.Image
		}
	case config.TypeAlbums:
		if i.Album != nil {
			return i.Album.Name, i.Album.Image
		}
	case config.TypeTracks:
		if i.Track != nil {
			image := i.Track.Image
			if image == "" && len(i.Track.Albums) > 0 {
				image = i.Track.Albums[0].Image
			}

			return i.Track.Name, image
		}
	}

	return "", ""
}
