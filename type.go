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

func (t Type) InstanceBuilder() *StructInstanceBuilder {
	return &StructInstanceBuilder{type_: t}
}

type StructInstanceFieldInit struct {
	Name  Generator
	Value Generator
}

func (i StructInstanceFieldInit) Generate() *jen.Statement {
	return i.Name.Generate().Op(":").Add(i.Value.Generate())
}

type StructInstanceBuilder struct {
	type_     Type
	Fields    []Generator
	MultiLine bool
}

func (b *StructInstanceBuilder) AppendField(f Generator) {
	b.Fields = append(b.Fields, f)
}

func (b *StructInstanceBuilder) AddNamedFieldString(name string, value Generator) {
	b.AppendField(StructInstanceFieldInit{Id(name), value})
}

func (b *StructInstanceBuilder) AddNamedField(name Generator, value Generator) {
	b.AppendField(StructInstanceFieldInit{name, value})
}

func (b *StructInstanceBuilder) Generate() *jen.Statement {
	var fields []Generator
	if !b.MultiLine {
		fields = b.Fields
	} else {
		l := len(b.Fields)
		fields = make([]Generator, l+1)
		for i, f := range b.Fields {
			fields[i] = Raw(jen.Line().Add(f.Generate()))
		}
		fields[l] = Line
	}
	return b.type_.Generate().Values(ToJenCodes(fields)...)
}
