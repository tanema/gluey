# Gluey

[![Go Reference](https://pkg.go.dev/badge/github.com/tanema/gluey.svg)](https://pkg.go.dev/github.com/tanema/gluey)

> Graphical Line User Experience Yes

Gluey is an opinionated graphic input library for CLI applications. It aims to
have a simple cross platform interface over configuability. It is meant to be a
port of [CLI::UI](https://github.com/shopify/cli-ui) for Go.

# Example

A simple example of how easy it is to use

```go
ctx := gluey.New()
// required text input
username, err := ctx.Ask("Username")
// Hidden output
passwd, err := ctx.Password("Password")
// confirm with default to true
agree, err := ctx.ConfirmDefault("Do you agree to our terms", true)
// Select Single
id, hardness, err := ctx.Select(
  "How hard is it to use this library?",
  []string{"Easy", "Medium", "Hard", "WTF"},
)
// Select Many
ids, editors, err := ctx.SelectMultiple(
  "Which Text Editors do you use",
  []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom", "other"},
)
```
