package generators

import (
	"github.com/dave/jennifer/jen"
)

// Value is a wrapper on top of [Generator] to provide easy access to type
// generation.
type Type struct{ Generator }

// NewType creates a Type representing a type in the local package with the
// specified name.
func NewType(name string) Type { return Type{Raw(jen.Id(name))} }

// NewTypePackage creates a Type represing a type imported from a package. The
// name is the type name, and pkg is the fully qualified package name.
//
// The local package alias is automatically created based on the import
// specifications of the generated file.
func NewTypePackage(name string, pkg string) Type { return Type{Raw(jen.Qual(pkg, name))} }
func (t Type) Pointer() Generator                 { return Raw(jen.Op("*").Add(t.Generate())) }

func (t Type) TypeParam(g Generator) Value {
	return Value{Raw(t.Generate().Index(g.Generate()))}
}

func (t Type) CreateInstance(values ...Generator) Value {
	return Value{Raw(t.Generate().Values(ToJenCodes(values)...))}
}

func (t Type) Literal(options ...func(*StructLiteral)) *StructLiteral {
	return &StructLiteral{Type: t}
}

// StructLiteralKeyElement generates an element with a key in a struct literal
type StructLiteralKeyElement struct {
	Key   Generator
	Value Generator
}

func (i StructLiteralKeyElement) Generate() *jen.Statement {
	return i.Key.Generate().Op(":").Add(i.Value.Generate())
}

// StructLiteral builds struct literals, i.e., statement that creates a struct
// and initialises it with values.
//
// Type should be a Generator that generates an
// identifier. Elements contain the values for all the elements.
//
// [KeyField] or [StructLiteralInstanceFieldInit] can be used to create a
// generator for an element with a key.
type StructLiteral struct {
	Type      Generator
	Elements  []Generator
	MultiLine bool
}

// Creates an element in the struct literal. Passing a generator that generates
// an expression will create a field without a key.
// [StructLiteral.KeyField] helps creating a field with a key.
func (b *StructLiteral) Field(f Generator) {
	b.Elements = append(b.Elements, f)
}

// Creates an element with a key in the struct literal
func (b *StructLiteral) KeyField(name Generator, value Generator) {
	b.Field(StructLiteralKeyElement{name, value})
}

func (b *StructLiteral) Generate() *jen.Statement {
	var fields []Generator
	if !b.MultiLine {
		fields = b.Elements
	} else {
		l := len(b.Elements)
		fields = make([]Generator, l+1)
		for i, f := range b.Elements {
			fields[i] = Raw(jen.Line().Add(f.Generate()))
		}
		fields[l] = Line
	}
	return b.Type.Generate().Values(ToJenCodes(fields)...)
}

func (b *StructLiteral) Value() Value {
	return Value{b}
}
