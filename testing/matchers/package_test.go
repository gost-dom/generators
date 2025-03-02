package matchers_test

import (
	"testing"

	gen "github.com/gost-dom/generators"
	. "github.com/gost-dom/generators/testing/matchers"

	"github.com/onsi/gomega"
)

func TestHaveRendered(t *testing.T) {
	g := gomega.NewWithT(t)
	x := gen.NewValue("x")
	expr := gen.Assign(x, gen.Lit(42))
	g.Expect(HaveRendered("x := 42").Match(expr)).To(gomega.BeTrue())
	g.Expect(HaveRendered("y := 42").Match(expr)).To(gomega.BeFalse())
	g.Expect(HaveRendered("x := 43").Match(expr)).To(gomega.BeFalse())

	compoundExpr := gen.StatementList(expr, gen.Return(x.Field("String").Call()))
	g.Expect(HaveRendered("x := 42").Match(compoundExpr)).To(gomega.BeFalse())
	g.Expect(HaveRendered("x := 42\nreturn x.String()").Match(compoundExpr)).To(gomega.BeTrue())
	g.Expect(HaveRenderedSubstring("x := 42").Match(compoundExpr)).To(gomega.BeTrue())
	g.Expect(HaveRendered(gomega.ContainSubstring("x := 42")).Match(compoundExpr)).
		To(gomega.BeTrue())
	g.Expect(HaveRendered(gomega.MatchRegexp(`x := \d{2}`)).Match(compoundExpr)).
		To(gomega.BeTrue())
	g.Expect(HaveRendered(gomega.MatchRegexp(`x := \d{3}`)).Match(compoundExpr)).
		To(gomega.BeFalse())
}
