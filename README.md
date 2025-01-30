# Generators

This package is part of the [Gost](https://github.com/gost-dom) to
support code generation from web specification IDL files.

However, the code is of general use, and exposed publicly.

> [!NOTE]
>
> Code is poorly documented, I will improve as I work more with this.

## Build on top of Jennifer

This library is a wrapper on top of [Jennifer](https://github.com/dave/jennifer)

Jennifer is a comprehensive library for generating go code generators, but there
are design decisions I would have made differently

- Expose a more declarative API
- Expose a higher level of abstraction

### Declarative vs. Imperative

This type of problem lends itself very well to a declarative composition. In
fact, this problem is very similar to building a UI. Modern web frameworks
generally compose high level components of low level components, each component
being a function of state.

Jennifer often mutatas values, making individual values difficult to reuse in
different contexts

### Level of abstraction

Jennifer has a model centered around the output tokens generated in code, e.g.
the function `Func` generating the `func` keyword. But the keword has different
uses.

- Declaring a function type
- Declaring a function literal

Likewise, `*` is created using `Op("*")`, but this has multiple uses as well:

- Declaring a pointer type
- Dereferencing a pointer variable

And those are just two examples.

This library tries to focus on what the code should do, not the tokens in the
generated source file. It uses abstractions like "Variable assignment",
"Pointer", "Reference", "Equals", rather than `Op(":=")`, `Op("*")`, `Op("&")`,
`Op("==")`.

If `t` is a `Type` representing a type in code, `t.Pointer()` represents the
pointer type to that type. If `v` is a `Value` representing an value, e.g., an
identifier, a struct literal, `t.Reference()` generates to to get a reference to
the value.


## Examples

This is not a full documentation, just to get you started.

Everything in this library is a `Generator`, and interface representing the
`Generate` function that can return a `*jen.Statement`.

```go
type Generator interface {
	Generate() *jen.Statement
}
```

### Types and Values

Two types, `Type` and `Value` are simple wrappers on top of `Generator` to
provide easy access to constructs 

E.g., on a `Value`, you can access fields, or call them. On a 

```Go
v := g.NewValue("MyStruct")
return StatementList(
    g.Assign(v, g.NewValue("NewMyStruct").Call()),
    g.Field("Initialize").Call(),
    g.Return(v),
    )
```

### Conditions

Simple `if`/`else` is implemented by `IfStmt`. `Eq` and `Neq` provides `==` and
`!=` support. `Gte`, greater-than or equal, is `>=`

```Go
IfStmt{
    Condition: Eq{ Lhs: value1, Rhs: Value2 },
    Block: someFunctionValue.Call(),
    // Else is optional
    Else: someOtherFunctionValue.Call(),
}
```


> [!NOTE]
>
> If you create custom support for new constructs, please create a PR.

### Creating a file

This package does not handle the overall file creation, package specification,
and import aliases. So you need to use Jen directly here

```go
func WriteGenerator(g Generator, w io.Writer) (error) {
    // Fully qualified package path
    file := jen.NewFilePath("example.com/my/package")
    file.HeaderComment("This file is generated. Do not edit.")
    // Potentially, create aliases for imports
    file.ImportAlias("github.com/tommie/v8go", "v8")
    file.Add(generator.Generate())
    return file.Render(w)
}
```

## It's a transparent layer (with an escape hatch)

This library does not try hide Jennifer, and the two can easily be intermixed.

Want to use a generator with Jennifer? Just call `Generate()`. To use a jennifer
value as a Generator, you can use the function `Raw`.

```
fooAsJenStmt := jen.Id("foo")
fooAsGenerator := generators.Raw(j)
fooRefAsJen := jen.Op("&").Add(fooAsGenerator.Generate())
```

## Philosophy: Embrace composition

The philosophy is to be able to compose larger structures out of individual
parts. Each level of composition adds a higher level of abstraction.

As an example

- Compose high level file generators from application specific generators
- Compose application specific generators from high level general purpose generators
- Compose high level general purpose generators from low-level general purpose
generators.

### High-level file generators

At the highest level, compose that parts that need to be in the file:

```Go
func FileContents() g.Generator {
    return StatementList(
        TypeGenerator(),
        ConstructorGenerator(),
        MethodsGenerator(),
        )
}
```

### High-level application specific generators

This would be generators created specifically for the types that exist in your
application:

```go
type MyTypeInstance struct {
    g.Generator
}

func (i MyTypeInstance) CallMethod1(arg1 Generator, arg2 Generator) Generator {
    v := g.Value{i.Generator}
    return v.Field("Method1").Call(arg1, arg2)
}

func Body() g.Generator {
    v := MyTypeInstance{g.NewValue("t")}
    return StatementList {
        Assign(v, g.Value("NewMyType").Call()),
        v.CallMethod1(g.Lit("foo"), g.Lit("bar")),
        g.Return(v),
    }
}
```

### High-level general purpose generators

A high-level general purpose generator could be an "Attribute", retrieving a private
read-only field:

```Go
type Receiver struct {
	Name g.Generator
	Type g.Generator
}

type Attribute struct {
	Name string
	Type g.Generator
	Receiver      Receiver
	ReadOnly      bool
}

func (a Attribute) Generate() *jen.Statement {
	field := g.ValueOf(a.Receiver.Name).Field(a.Name)
	getter := g.FunctionDefinition{
		Receiver: g.FunctionArgument(a.Receiver),
		Name:     upperCaseFirstLetter(a.Name),
		RtnTypes: g.List(a.Type),
		Body:     g.Return(field),
	}
	l := g.StatementList(
		getter,
		g.Line,
	)
	if !a.ReadOnly {
		argument := g.NewValue("val")
		l.Append(g.FunctionDefinition{
			Receiver: getter.Receiver,
			Name:     fmt.Sprintf("Set%s", getter.Name),
			Args:     g.Arg(argument, a.Type),
			Body:     g.Reassign(field, argument),
		})
	}
	return l.Generate()
}
```

From this, you can easily generate multiple, e.g. from a list of names:

```go
func GenerateAttribute(names []string) Generator {
	r := Receiver{
		Name: g.Id("t"),
		Type: g.Id("MyType"),
	}
	gg := make([]Generator, len(names))
	for i, n := range names {
		gg[i] = Attribute {
			Receiver: Receiver{
				Name: g.Id("t"),
				Type: g.Id("MyType"),
			},
			Name: g.Id(n),
			Type: g.Id("string"), // or: getTypeForAttribute(n)
		}
	}
	return StatementList(gg...)
}
```

I don't intend to add those kinds of types in this package, but could eventually
appear in some kind of _support_ package.

### Low-level general purpose generators

From the previous example, you could have extracted, e.g. just a Getter and a
Setter. Or AssignField, or ...
