package jujusvg

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/charm/v7"
)

func TestLinkFetchIcons(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	tests := map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~charming-devs/precise/elasticsearch-2.svg" />
			</svg>`),
		"~juju-jitsu/precise/charmworld-58": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~juju-jitsu/precise/charmworld-58.svg" />
			</svg>`),
		"precise/mongodb-21": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/precise/mongodb-21.svg" />
			</svg>`),
	}
	iconURL := func(_ context.Context, ref *charm.URL) (string, error) {
		return "/" + ref.Path() + ".svg", nil
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, qt.IsNil)
	err = b.Verify(nil, nil, nil)
	c.Assert(err, qt.IsNil)
	fetcher := LinkFetcher{
		IconURL: iconURL,
	}
	iconMap, err := fetcher.FetchIcons(ctx, b)
	c.Assert(err, qt.IsNil)
	for charm, link := range tests {
		assertXMLEqual(c, []byte(iconMap[charm]), []byte(link))
	}
}

func TestHTTPFetchIcons(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	fetchCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		fmt.Fprintln(w, fmt.Sprintf("<svg>%s</svg>", r.URL.Path))
	}))
	defer ts.Close()

	tsIconURL := func(_ context.Context, ref *charm.URL) (string, error) {
		return ts.URL + "/" + ref.Path() + ".svg", nil
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, qt.IsNil)
	err = b.Verify(nil, nil, nil)
	c.Assert(err, qt.IsNil)
	// Only one copy of precise/mongodb-21
	b.Applications["duplicateApplication"] = &charm.ApplicationSpec{
		Charm:    "cs:precise/mongodb-21",
		NumUnits: 1,
	}
	fetcher := HTTPFetcher{
		Concurrency: 1,
		IconURL:     tsIconURL,
	}
	iconMap, err := fetcher.FetchIcons(ctx, b)
	c.Assert(err, qt.IsNil)
	c.Assert(iconMap, qt.DeepEquals, map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte("<svg>/~charming-devs/precise/elasticsearch-2.svg</svg>\n"),
		"~juju-jitsu/precise/charmworld-58":      []byte("<svg>/~juju-jitsu/precise/charmworld-58.svg</svg>\n"),
		"precise/mongodb-21":                     []byte("<svg>/precise/mongodb-21.svg</svg>\n"),
	})

	fetcher.Concurrency = 10
	iconMap, err = fetcher.FetchIcons(ctx, b)
	c.Assert(err, qt.IsNil)
	c.Assert(iconMap, qt.DeepEquals, map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte("<svg>/~charming-devs/precise/elasticsearch-2.svg</svg>\n"),
		"~juju-jitsu/precise/charmworld-58":      []byte("<svg>/~juju-jitsu/precise/charmworld-58.svg</svg>\n"),
		"precise/mongodb-21":                     []byte("<svg>/precise/mongodb-21.svg</svg>\n"),
	})
	c.Assert(fetchCount, qt.Equals, 6)
}

func TestHTTPBadIconURL(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad-wolf", http.StatusForbidden)
		return
	}))
	defer ts.Close()

	tsIconURL := func(_ context.Context, ref *charm.URL) (string, error) {
		return ts.URL + "/" + ref.Path() + ".svg", nil
	}

	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, qt.IsNil)
	err = b.Verify(nil, nil, nil)
	c.Assert(err, qt.IsNil)
	fetcher := HTTPFetcher{
		Concurrency: 1,
		IconURL:     tsIconURL,
	}
	iconMap, err := fetcher.FetchIcons(ctx, b)
	c.Assert(err, qt.ErrorMatches, fmt.Sprintf("cannot retrieve icon from %s.+\\.svg: 403 Forbidden.*", ts.URL))
	c.Assert(iconMap, qt.IsNil)

	fetcher.Concurrency = 10
	iconMap, err = fetcher.FetchIcons(ctx, b)
	c.Assert(err, qt.ErrorMatches, fmt.Sprintf("cannot retrieve icon from %s.+\\.svg: 403 Forbidden.*", ts.URL))
	c.Assert(iconMap, qt.IsNil)
}
