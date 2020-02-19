package main

import (
	"testing"

	glob "github.com/pachyderm/ohmyglob"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	pattern string
	assert  func(g *glob.Glob, t *testing.T)
}

func TestGlob(t *testing.T) {
	var g *glob.Glob

	cases := []TestCase{
		{
			pattern: "*.js",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.True(t, g.Match("a.js"))
				assert.True(t, g.Match("b.js"))
				assert.True(t, g.Match(".js"))
				assert.False(t, g.Match("b.u"))
			},
		},

		{
			pattern: "a?c",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.True(t, g.Match("abc"))
				assert.True(t, g.Match("aac"))
				assert.False(t, g.Match("ac"))
				assert.False(t, g.Match("ucc"))
			},
		},

		{
			pattern: "[a-z]b",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.True(t, g.Match("ab"))
				assert.True(t, g.Match("bb"))
				assert.False(t, g.Match("aab"))
				assert.False(t, g.Match("0b"))
			},
		},

		{
			pattern: "[^a-z]b",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("ab"))
				assert.False(t, g.Match("bb"))
				assert.True(t, g.Match("0b"))
			},
		},

		{
			pattern: "!(a|b|c)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.True(t, g.Match("u"))
				assert.True(t, g.Match("aa"))
				assert.False(t, g.Match("a"))
			},
		},

		{
			pattern: "?(a|b|c)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("u"))
				assert.True(t, g.Match("a"))
				assert.True(t, g.Match("b"))
				assert.True(t, g.Match(""))
			},
		},

		{
			pattern: "+(a|b|c)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("u"))
				assert.True(t, g.Match("a"))
				assert.True(t, g.Match("aaaa"))
				assert.True(t, g.Match("aabbcc"))
				assert.False(t, g.Match(""))
			},
		},

		{
			pattern: "*(a|b|c)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("u"))
				assert.True(t, g.Match("a"))
				assert.True(t, g.Match("aaaa"))
				assert.True(t, g.Match(""))
			},
		},

		{
			pattern: "@(aaa|bbb)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("uuu"))
				assert.False(t, g.Match("a"))
				assert.True(t, g.Match("aaa"))
				assert.True(t, g.Match("bbb"))
			},
		},

		{
			pattern: "**",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.True(t, g.Match("a"))
				assert.True(t, g.Match("a/b"))
				assert.True(t, g.Match("a/b/c"))
			},
		},

		{
			pattern: "**/*.js",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("file.js"))
				assert.True(t, g.Match("a/file.js"))
				assert.True(t, g.Match("a/b/file.js"))
				assert.True(t, g.Match("a/b/c/file.js"))
			},
		},

		{
			pattern: "**/!(a|b)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("file"))
				assert.True(t, g.Match("a/file"))
				assert.True(t, g.Match("a/b/file"))
				assert.False(t, g.Match("a/a"))
				assert.False(t, g.Match("a/b"))
			},
		},

		{
			pattern: "**/!(*.common.*|*.html)",
			assert: func(g *glob.Glob, t *testing.T) {
				assert.False(t, g.Match("file"))
				assert.True(t, g.Match("a/file"))
				assert.True(t, g.Match("a/b/file"))
				assert.True(t, g.Match("a/b/file.js"))
				assert.False(t, g.Match("a/file.common.js"))
				assert.False(t, g.Match("a/file.html"))
			},
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable

		t.Run(tc.pattern, func(t *testing.T) {
			t.Parallel()
			g = glob.MustCompile(tc.pattern)
			tc.assert(g, t)
		})

	}

}
