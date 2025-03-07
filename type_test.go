package generators_test

import (
	"testing"

	. "github.com/gost-dom/generators"
	"github.com/gost-dom/generators/testing/matchers"
	"github.com/onsi/gomega"
)

func TestGenerateStructInstanceWithNamedFields(t *testing.T) {
	g := gomega.NewWithT(t)
	b := NewType("MyType").Literal()
	b.KeyField(Id("StringField"), Lit("string value"))
	b.KeyField(Id("IntField"), Lit(42))

	g.Expect(b).To(matchers.HaveRendered(`MyType{StringField: "string value", IntField: 42}`))

	bRef := b.Value().Reference()
	g.Expect(bRef).To(matchers.HaveRendered(`&MyType{StringField: "string value", IntField: 42}`))

	b.MultiLine = true
	g.Expect(b).To(matchers.HaveRendered(`MyType{
	StringField: "string value",
	IntField:    42,
}`))
}
