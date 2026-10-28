package fonts

import (
	"slices"
	"testing"
)

// TestImportURLs verifies how ImportURLs resolves tracked font families, the default font and
// mj-font declarations into one ordered list of stylesheet URLs, including the cases MJML
// templates can't express: entries added to GoogleFontsMapping, aliases sharing a URL, and
// empty URLs.
func TestImportURLs(t *testing.T) {
	const (
		googleLato   = "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"
		googleRoboto = "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
		googleUbuntu = "https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700"
		googleMont   = "https://fonts.googleapis.com/css?family=Montserrat:300,400,500,700"
		customLato   = "https://fonts.example.com/lato.css"
		customRoboto = "https://fonts.example.com/roboto.css"
		customUbuntu = "https://fonts.example.com/ubuntu.css"
		comicNeue    = "https://fonts.example.com/comic-neue.css"
		raleway      = "https://fonts.example.com/raleway.css"
		pacifico     = "https://fonts.example.com/pacifico.css"
	)

	tests := []struct {
		name          string
		addedEntries  map[string]string
		fontFamilies  []string
		declarations  []Declaration
		importDefault bool
		want          []string
	}{
		{
			name:         "Tracked fonts in detection order",
			fontFamilies: []string{"Roboto", "Lato, Arial"},
			want:         []string{googleLato, googleRoboto},
		},
		{
			name: "Nothing to import",
			want: nil,
		},
		{
			name:          "Default font",
			importDefault: true,
			want:          []string{googleUbuntu},
		},
		{
			name:          "Default font at its position",
			fontFamilies:  []string{"Montserrat"},
			importDefault: true,
			want:          []string{googleUbuntu, googleMont},
		},
		{
			name:         "Declaration replaces a built-in URL at its position",
			fontFamilies: []string{"Roboto", "Lato"},
			declarations: []Declaration{{Name: "Lato", Href: customLato}},
			want:         []string{customLato, googleRoboto},
		},
		{
			name:          "Declaration replaces the default font",
			declarations:  []Declaration{{Name: "Ubuntu", Href: customUbuntu}},
			importDefault: true,
			want:          []string{customUbuntu},
		},
		{
			name:         "Declared name imported though no tracked font uses it",
			declarations: []Declaration{{Name: "Raleway", Href: raleway}},
			want:         []string{raleway},
		},
		{
			name:         "Declared names follow built-in and added entries",
			addedEntries: map[string]string{"Comic Neue": comicNeue},
			fontFamilies: []string{"Comic Neue", "Raleway", "Lato"},
			declarations: []Declaration{{Name: "Raleway", Href: raleway}},
			want:         []string{googleLato, comicNeue, raleway},
		},
		{
			name:         "Names match exactly",
			fontFamilies: []string{"Lato"},
			declarations: []Declaration{{Name: "lato", Href: customLato}},
			want:         []string{googleLato, customLato},
		},
		{
			name: "Redeclared name imports its last href at its first position",
			declarations: []Declaration{
				{Name: "Raleway", Href: "https://fonts.example.com/raleway-old.css"},
				{Name: "Pacifico", Href: pacifico},
				{Name: "Raleway", Href: raleway},
			},
			want: []string{raleway, pacifico},
		},
		{
			name:         "Declaration with an empty href ignored",
			fontFamilies: []string{"Lato"},
			declarations: []Declaration{{Name: "Lato", Href: customLato}, {Name: "Lato", Href: ""}, {Name: "Raleway", Href: ""}},
			want:         []string{customLato},
		},
		{
			name:         "Alias keeps its URL when the name it aliases is declared",
			addedEntries: map[string]string{"Inter": googleRoboto},
			fontFamilies: []string{"Inter"},
			declarations: []Declaration{{Name: "Roboto", Href: customRoboto}},
			want:         []string{customRoboto, googleRoboto},
		},
		{
			name:         "URLs deduplicated",
			fontFamilies: []string{"Lato"},
			declarations: []Declaration{{Name: "Raleway", Href: googleLato}},
			want:         []string{googleLato},
		},
		{
			name:         "Empty URLs skipped",
			addedEntries: map[string]string{"Empty Font": ""},
			fontFamilies: []string{"Empty Font"},
			want:         nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for name, url := range tc.addedEntries {
				if _, exists := GoogleFontsMapping[name]; exists {
					t.Fatalf("test fixture %q unexpectedly already present in GoogleFontsMapping", name)
				}
				t.Cleanup(func() { delete(GoogleFontsMapping, name) })
				GoogleFontsMapping[name] = url
			}

			got := ImportURLs(tc.fontFamilies, tc.declarations, tc.importDefault)
			if !slices.Equal(got, tc.want) {
				t.Errorf("ImportURLs() = %q, want %q", got, tc.want)
			}
		})
	}
}
