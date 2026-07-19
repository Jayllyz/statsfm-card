package statsfm

import "testing"

func TestItemNameAndImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		itemType  string
		item      Item
		wantName  string
		wantImage string
	}{
		{
			name:      "artist",
			itemType:  "artists",
			item:      Item{Artist: &Entity{Name: "Muse", Image: "muse.png"}},
			wantName:  "Muse",
			wantImage: "muse.png",
		},
		{
			name:      "album",
			itemType:  "albums",
			item:      Item{Album: &Entity{Name: "Origin of Symmetry", Image: "origin.png"}},
			wantName:  "Origin of Symmetry",
			wantImage: "origin.png",
		},
		{
			name:      "track with own image",
			itemType:  "tracks",
			item:      Item{Track: &Track{Name: "Hysteria", Image: "hysteria.png"}},
			wantName:  "Hysteria",
			wantImage: "hysteria.png",
		},
		{
			name:      "track falls back to album image",
			itemType:  "tracks",
			item:      Item{Track: &Track{Name: "Hysteria", Albums: []Entity{{Image: "album.png"}}}},
			wantName:  "Hysteria",
			wantImage: "album.png",
		},
		{
			name:      "track without any image",
			itemType:  "tracks",
			item:      Item{Track: &Track{Name: "Hysteria"}},
			wantName:  "Hysteria",
			wantImage: "",
		},
		{
			name:      "missing entity for type",
			itemType:  "artists",
			item:      Item{},
			wantName:  "",
			wantImage: "",
		},
		{
			name:      "unknown type",
			itemType:  "playlists",
			item:      Item{Artist: &Entity{Name: "Muse", Image: "muse.png"}},
			wantName:  "",
			wantImage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotName, gotImage := tt.item.NameAndImage(tt.itemType)
			if gotName != tt.wantName || gotImage != tt.wantImage {
				t.Errorf("NameAndImage(%q) = (%q, %q), want (%q, %q)",
					tt.itemType, gotName, gotImage, tt.wantName, tt.wantImage)
			}
		})
	}
}
