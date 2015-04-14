package jujusvg

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v5-unstable"
)

type IconFetcherSuite struct{}

var _ = gc.Suite(&IconFetcherSuite{})

func (s *IconFetcherSuite) TestLinkFetchIcons(c *gc.C) {
	tests := map[string]string{
		"~charming-devs/precise/elasticsearch-2": `
			<svg xlmns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~charming-devs/precise/elasticsearch-2.svg" />
			</svg>`,
		"~juju-jitsu/precise/charmworld-58": `
			<svg xlmns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~juju-jitsu/precise/charmworld-58.svg" />
			</svg>`,
		"precise/mongodb-21": `
			<svg xlmns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/precise/mongodb-21.svg" />
			</svg>`,
	}
	iconUrl := func(ref *charm.Reference) string {
		return "/" + ref.Path() + ".svg"
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil)
	c.Assert(err, gc.IsNil)
	fetcher := LinkFetcher{
		IconURL: iconUrl,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	for charm, link := range tests {
		assertXMLEqual(c, []byte(iconMap[charm]), []byte(link))
	}
}

func (s *IconFetcherSuite) TestHttpFetchIcons(c *gc.C) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<svg></svg>")
	}))
	defer ts.Close()

	tsIconUrl := func(ref *charm.Reference) string {
		return ts.URL + "/" + ref.Path() + ".svg"
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil)
	c.Assert(err, gc.IsNil)
	fetcher := HttpFetcher{
		FetchConcurrently: false,
		IconURL:           tsIconUrl,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string]string{
		"~charming-devs/precise/elasticsearch-2": "<svg></svg>\n",
		"~juju-jitsu/precise/charmworld-58":      "<svg></svg>\n",
		"precise/mongodb-21":                     "<svg></svg>\n",
	})

	fetcher.FetchConcurrently = true
	iconMap, err = fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string]string{
		"~charming-devs/precise/elasticsearch-2": "<svg></svg>\n",
		"~juju-jitsu/precise/charmworld-58":      "<svg></svg>\n",
		"precise/mongodb-21":                     "<svg></svg>\n",
	})
}

func (s *IconFetcherSuite) TestHttpBadIconURL(c *gc.C) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad-wolf", http.StatusForbidden)
		return
	}))
	defer ts.Close()

	tsIconUrl := func(ref *charm.Reference) string {
		return ts.URL + "/" + ref.Path() + ".svg"
	}

	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil)
	c.Assert(err, gc.IsNil)
	fetcher := HttpFetcher{
		FetchConcurrently: false,
		IconURL:           tsIconUrl,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("Error retrieving icon from %s.+\\.svg: 403 Forbidden", ts.URL))
	c.Assert(iconMap, gc.IsNil)

	fetcher.FetchConcurrently = true
	iconMap, err = fetcher.FetchIcons(b)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("Error retrieving icon from %s.+\\.svg: 403 Forbidden", ts.URL))
	c.Assert(iconMap, gc.IsNil)
}
