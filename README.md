# go-dom - Headless browser for Go

> [!NOTE] 
>
> This is only a POC
>
> Much of this readme file was written before any code was added, and describes
> the intent of the project, and the benefits that it brings.
>
> This starts as a hobby project for learning; there is no guarantee that it
> will ever become a useful tool; unless it will gain traction and support of
> multiple developers.

[Software license](./LICENSE.txt) (note: I currently distribute the source under
the MIT license, but as this will most likely depend on v8go, I need to verify
that the license is compatible).

While the SPA[^1] dominates the web today, some applications still render
server-side HTML, and HTMX is gaining in popularity. Go has some popularity as a
back-end language for HTMX.

In Go, writing tests for the HTTP handler is easy if all you need to do is
verify the response.

But if you need to test at a higher lever, for example verify how any JavaScript
code effects the page; you would need to use browser automation, like
[Selenium](https://www.selenium.dev/), and this introduces a significant 
overhead; not only from out-of-process communication with the browser, but also
the necessity of launching your server.

This overhead discourages a TDD loop.

The purpose of this project is to support a TDD feedback loop for code
delivering HTML, and where merely verifying the HTTP response isn't enough, but
you want to verify:

- JavaScript code has the desired behaviour
- General browser behaviour is verified, e.g. 
  - clicking a `<button type="submit">` submits the form
  - A and a redirect response is followed.

Some advantages of a native headless browser are:

- No need to wait for a browser to launch.
- Everything works in-process, so interacting with the browser from test does
  not incur the overhead of out-of-process communication, and you could for
  example redirect all console output to go code easily.
- You can request application directly through the 
  [`http.Handler`](https://pkg.go.dev/net/http#Handler); so no need to start an
  HTTP server.
- You can run parallel tests in isolation as each can create their own _instance_
  of the HTTP handler.[^2]

Some disadvantages compared to e.g. Selenium.

- You cannot verify how it look; e.g. you cannot get a screenshot of a failing test
  - This means you cannot create snap-shot tests detect undesired UI changes.[^3]
- You cannot verify that everything works in _all supported browsers_.

This isn't intended as a replacement for the cases where an end-2-end test is
the right choice. It is intended as a tool to help when you want a smaller
isolated test, e.g. mocking out part of the behaviour; but 

## Project status

The current state of the project is, you can start a "browser" connected to a Go
`http.Handler` (the intention is to use the root handler, that you normally
expose on a TCP port; not a route, but feel free to do whatever you want).

The browser can open an HTML file, execute included scripts (remote scripts are
downloaded). The DOM is exposed as native Go objects, allowing the test code to
inspect or modify them, and JavaScript can also be executed from Go code.

Only a very minimal subset of the DOM specification is implemented.



- HTML parsing is done in 2 steps
  - Step 1 parsing of HTML using [x/net/html](https://pkg.go.dev/golang.org/x/net/html)
      - Using `x/net/html` gives HTML rendering (i.e., support for `outerHTML`) out
        of the box.
      - Libraries exist implementing XPath on top of this.
  - 2nd pass into my own structure
    - `x/net/html` does not have the interface that a Browser wants, so I wrap
    this to provide the browser DOM both to JavaScript and Go.
    - The library doesn't support the insertion steps, e.g., when a `<script>` is
      connected to the DOM, it should be executed (simplified).
- Embedding of v8 engine.[^4]
  - Expose the navite Go objects to JavaScript

### Memory Leaks

The current implementation is leaking memory for the scope of a "Browser". I.e.,
all DOM nodes created and deleted for the lifetime of the browser will stay in
memory until the browser is ready for garbage collection.

The problem here is that this is a marriage between two garbage collected
systems, and what is conceptually _one object_ is split into two, a Go object
and a JavaScript wrapper. As long of them is reachable; so must the other be.

I could join them into one; but that would result in an undesired coupling; the
DOM implementation being coupled to the JavaScript execution engine.

Another solution to this problem involves the use of weak references. This
exists as an `internal` but [was
accepted](https://github.com/golang/go/issues/67552) as a feature.

Because of that, and because the browser is only intended to be kept alive for
the scope of a single short lived test, I have postponed dealing with memory
management.

### Demonstration

An early example showing HTML being loaded with a script. Notice:

- The HTML parser creates a `<head>`, even if missing in the source.
- Whitespace is not inserted in the DOM outside the body (line break and
  indentation is only processed in the body.
- The script is executed when connected to the DOM; which is why it doesn't see
  the `<div>` element after, as well as whitespace.

```go
It("Runs the script when connected to DOM", func() {
    window := ctx.Window()
    window.SetScriptRunner(ctx)
    window.LoadHTML(`
<html>
  <body>
    <script>window.sut = document.documentElement.outerHTML</script>
    <div>I should not be in the output</div>
  </body>
</html>
`,
    )
    Expect(
        ctx.MustRunTestScript("window.sut"),
    ).To(Equal(`<html><head></head><body>
    <script>window.sut = document.documentElement.outerHTML</script></body></html>`))
})
```

### Next up


#### Cleanup

First a bit of cleanup. Necessary dependencies was identified in order to
execute scripts during parsing. Minimal code was written to get the test to
pass; but a bit messy.

#### Memory Management

This is an integration between two languages both with their own Garbage
collector; The same object lives in two worlds; but can at any time be reachable
in both; or just one. In the latter case, the object must be kept alive in the
other. When not reachable in either world, it should be ready for garbage
collection in both.

##### Keep JavaScript objects alive

JavaScript objects are created to wrap Go objects. Two JavaScript
values representing the same DOM object must always be equal. Therefore the Go
code needs to keep JavaScript objects alive, even when they are out of scope in
JavaScript. V8 has mechanisms for controlling this.

##### Keep Go objects alive

But also, A Go object may have run out of scope, e.g. an element was
disconnected from the DOM. But a JavaScript object may still have a reference to
this object. So JavaScript code needs to keep Go objects alive; but when the
JavaScript object goes out of scope, the Go object should be ready for garbage
collection. V8 also has some support for this.

### Future goals

There is much to do, which includes (but this is not a full list):

- Support all DOM elements, including SVG elements and other namespaces.
- Handle bad HTML gracefully (browsers don't generate an error if an end tag is
  missing or misplaced)
- Implement all standard JavaScript classes that a browser should support; but
  not provided by the V8 engine.
  - JavaScript polyfills would be a good starting point, where some exist.
  - Conversion to native go implementations would be prioritized on usage, e.g.
    [`fetch`](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) 
    would be high in the list of priorities.
- Implement default browser behaviour for user interaction, e.g. pressing 
  <key>enter</key> when an input field has focus should submit the form.

### Long Term Goals

#### CSS Parsing

Parsing CSS woule be nice, allowing test code to verify the resulting styles of
an element; but having a working DOM with a JavaScript engine is higher
priority.

#### Mock external sites

The system may depend on external sites in the browser, most notably identity
providers (IDP), where your app redirects to the IDP, which redirects on
successful login; but could be other services such as map providers, etc.

For testing purposes, replacing this with a dummy replacement would have some
benefits:

- The verification of your system doesn't depend on the availability of an
  external service; when working offline
- Avoid tests breaking because of changes to the external system.
- For an identity provider
  - Avoid pollution of dummy accounts to run your test suite.
  - Avoid the locking the account due to _suspiscious activity_.
  - The IDP may use a Captcha or 2FA that can be impossible; or difficult to
    control from tests, and would cause a significant slowdown to the test
    suite.
- For applications like map providers
  - Avoid being billed for API use during testing.

## Help

This project will likely die without help. If you are interested in this, I
would welcome contributions. Particularly if:

- ~~You have experience building tokenisers and parsers, especially HTML.~~
  - After first building my own parser, I moved to `x/net/html`, which seems
    like the right choice; at least for now.
- You have intimate knowledge of Go's garbage collection mechanics.
  - If you don't have the time or desire to help _code_ on this project, ~~I would
    appreciate peer reviews on those parts of the code.~~
    - I have postponed solving that problem until Go gets weak references.
    - However, if you do see another solution to the leaking problem, let me
      know.
- You have _intimate knowledge_ of how the DOM works in the browser, and can 
  help detect poor design decisions early. For example:
  - should the internal implementation use `document.CreateElement()`
    when parsing HTML into DOM? 
    - Would it be a big mistake to do so? 
    - Would it be a big mistake to _not_ do so? 
    - Is is it a _doesn't matter_, whatever makes the code clean, issue?
  - Which "objects" should I expose from Go to v8? and where should the
    functions live? The objects themselves, or should I create prototype
    in Go code? (I think I _should_ make prototype objects)
- You have knowledge of the whatwg IDL, and what kind of code could be
  auto-generated from the IDL
- You have experience working with the v8 engine, particularly exposing internal
  objects to JavaScript (which is then External to JavaScript).
  - In particular, if you've done this from Go.

## Out of scope.

### Accessibility tree

It is not currently planned that this library should maintain the accessibility
tree; nor provide higher level testing capabilities like what
[Testing Library](https://testing-library.com) provides for JavaScript.

These problems _should_ eventually be solved, but could easily be implemented in
a different library with dependency to the DOM alone.

### Visual Rendering

It is not a goal to be able to provide a visual rendering of the DOM. 

But just like the accessibility tree, this could be implemented in a new library
depending only on the interface from here.


[^1]: Single-Page app
[^2]: This approach allows you to mock databases, and other external services;
A few integration tests that use a real database, message bus, or other external
services, is a good idea. Here, isolation of parallel tests may be
non-trivial; depending on the type of application.
[^3]: I generally dislike snapshot tests; as they don't _describe_ expected
behaviour, only that the outcome mustn't change. There are a few cases where
where snapshot tests are the right choice, but they should be avoided for a TDD
process.
[^4]: The engine is based on the v8go project by originally by @rogchap, later
kept up-to-date by @tommie; who did a remarkale job of automatically keeping the
v8 dependencies up-to-date. But many necessary features of V8 are not exported;
which I am adding in my own fork.
