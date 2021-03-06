package display

import (
	"errors"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// This file contains funcs that are small, self-contained utilities.

// isValidTab indicates whether the passed tab is still being used, even if it's not currently displayed.
func isValidTab(t *tab) bool {
	tempTabs := tabs
	for i := range tempTabs {
		if tempTabs[i] == t {
			return true
		}
	}
	return false
}

func leftMargin() int {
	return int(float64(termW) * viper.GetFloat64("a-general.left_margin"))
}

func textWidth() int {
	if termW <= 0 {
		// This prevent a flash of 1-column text on startup, when the terminal
		// width hasn't been initialized.
		return viper.GetInt("a-general.max_width")
	}

	rightMargin := leftMargin()
	if leftMargin() > 10 {
		// 10 is the max right margin
		rightMargin = 10
	}

	max := termW - leftMargin() - rightMargin
	if max < viper.GetInt("a-general.max_width") {
		return max
	}
	return viper.GetInt("a-general.max_width")
}

// queryEscape is the same as url.PathEscape, but it also replaces the +.
// This is because Gemini requires percent-escaping for queries.
func queryEscape(path string) string {
	return strings.ReplaceAll(url.PathEscape(path), "+", "%2B")
}

// resolveRelLink returns an absolute link for the given absolute link and relative one.
// It also returns an error if it could not resolve the links, which should be displayed
// to the user.
func resolveRelLink(t *tab, prev, next string) (string, error) {
	if !t.hasContent() {
		return next, nil
	}

	prevParsed, _ := url.Parse(prev)
	nextParsed, err := url.Parse(next)
	if err != nil {
		return "", errors.New("link URL could not be parsed")
	}
	return prevParsed.ResolveReference(nextParsed).String(), nil
}

// normalizeURL attempts to make URLs that are different strings
// but point to the same place all look the same.
//
// Example: gemini://gus.guru:1965/ and //gus.guru/.
// This function will take both output the same URL each time.
//
// The string passed must already be confirmed to be a URL.
// Detection of a search string vs. a URL must happen elsewhere.
//
// It only works with absolute URLs.
func normalizeURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	if !strings.Contains(u, "://") && !strings.HasPrefix(u, "//") {
		// No scheme at all in the URL
		parsed, err = url.Parse("gemini://" + u)
		if err != nil {
			return u
		}
	}
	if parsed.Scheme == "" {
		// Always add scheme
		parsed.Scheme = "gemini"
	} else if parsed.Scheme != "gemini" {
		// Not a gemini URL, nothing to do
		return u
	}

	parsed.User = nil    // No passwords in Gemini
	parsed.Fragment = "" // No fragments either
	if parsed.Port() == "1965" {
		// Always remove default port
		parsed.Host = parsed.Hostname()
	}

	// Add slash to the end of a URL with just a domain
	// gemini://example.com -> gemini://example.com/
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String()
}
