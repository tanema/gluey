# Gluey

> Graphical Line User Experience Yes

Gluey is an opinionated graphic input library for CLI applications. It aims to have a simple cross platform interface over configuability. It is meant to be a port of [CLI::UI](https://github.com/shopify/cli-ui) for Go.

# Example

A simple example of how easy it is to use

```
ctx := gluey.New()
ctx.Ask("Username") // required text input
ctx.Password("Password") // Hidden output
ctx.Confirm("Do you agree to our terms", true) // confirm with default to true
ctx.SelectMultiple("Which Text Editors do you use", []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom", "other"}) // Select Many
```

### Todo

- select
  - weird characters during select in windows cmd.exe
  - linebreak long options
