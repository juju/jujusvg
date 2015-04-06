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

var (
	ts      *httptest.Server
	iconUrl func(*charm.Reference) string
)

func (s *IconFetcherSuite) SetUpTest(c *gc.C) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<svg></svg>")
	}))

	iconUrl = func(ref *charm.Reference) string {
		return ts.URL + "/" + ref.Path() + ".svg"
	}
}

func (s *IconFetcherSuite) TearDownTest(c *gc.C) {
	ts.Close()
}

func (s *IconFetcherSuite) TestFetchIcons(c *gc.C) {
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil)
	c.Assert(err, gc.IsNil)
	fetcher := HttpFetcher{
		FetchConcurrently: false,
		IconURL:           iconUrl,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string]string{
		"~charming-devs/precise/elasticsearch-2": "<svg></svg>\n",
		"~juju-jitsu/precise/charmworld-58":      "<svg></svg>\n",
		"precise/mongodb-21":                     "<svg></svg>\n",
	})
}

func (s *IconFetcherSuite) TestFetchIconsConcurrent(c *gc.C) {
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil)
	c.Assert(err, gc.IsNil)
	fetcher := HttpFetcher{
		FetchConcurrently: true,
		IconURL:           iconUrl,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string]string{
		"~charming-devs/precise/elasticsearch-2": "<svg></svg>\n",
		"~juju-jitsu/precise/charmworld-58":      "<svg></svg>\n",
		"precise/mongodb-21":                     "<svg></svg>\n",
	})
}
