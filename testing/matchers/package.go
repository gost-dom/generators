// Package matchers contain matchers for use with the [gomega] assertion library.
package matchers

import (
	"bytes"

	"github.com/gost-dom/generators"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func render(g generators.Generator) (string, error) {
	var b bytes.Buffer
	err := g.Generate().Render(&b)
	return b.String(), err
}

// HaveRendered returns a [types.GomegaMatcher] that matches string rendered
// statement against another GomegaMatcher. If the expected value is not a
// GomegaMatcher, it will implicitly match for equality using [gomega.Equal].
func HaveRendered(expected interface{}) types.GomegaMatcher {
	matcher, ok := expected.(types.GomegaMatcher)
	if !ok {
		return HaveRendered(gomega.Equal(expected))
	}
	return gomega.WithTransform(render, matcher)
}

// HaveRenderedSubstring is simple helper for [HaveRendered] for a very common
// use case to test for the presence of a substrubg in the rendered statement.
// Short for HaveRendered(gomega.ContainSubstring(expected)).
func HaveRenderedSubstring(expected string) types.GomegaMatcher {
	return HaveRendered(gomega.ContainSubstring(expected))
}
