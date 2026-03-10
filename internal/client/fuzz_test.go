package client

import "testing"

func FuzzParseLinkNext(f *testing.F) {
	f.Add("")
	f.Add(`<https://api.example.com/page2>; rel="next"`)
	f.Add(`<https://api.example.com/page1>; rel="prev", <https://api.example.com/page3>; rel="next"`)
	f.Add(`<https://api.example.com/page1>; rel="prev"`)
	f.Add(`<>; rel="next"`)
	f.Add(`malformed header`)

	f.Fuzz(func(t *testing.T, header string) {
		parseLinkNext(header) // must not panic
	})
}

func FuzzParseRetryAfter(f *testing.F) {
	f.Add("")
	f.Add("5")
	f.Add("0")
	f.Add("-1")
	f.Add("999999999")
	f.Add("not-a-number")
	f.Add("Wed, 21 Oct 2015 07:28:00 GMT")

	f.Fuzz(func(t *testing.T, value string) {
		d := parseRetryAfter(value)
		if d < 0 {
			t.Errorf("parseRetryAfter(%q) returned negative duration: %v", value, d)
		}
	})
}

func FuzzParsePage(f *testing.F) {
	f.Add("")
	f.Add("https://api.example.com/cards.json?page=2")
	f.Add("https://api.example.com/cards.json")
	f.Add("not-a-url")
	f.Add("?page=abc")

	f.Fuzz(func(t *testing.T, nextURL string) {
		ParsePage(nextURL) // must not panic
	})
}
