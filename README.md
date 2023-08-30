# 🛡 arguard

Linter for Go that checks static arguments to function agains function [guards](https://en.wikipedia.org/wiki/Guard_(computer_science)) (aka contracts).

## Example

Let's say, you have the following function:

```go
func div(n, d float64) float64 {
  if in == 0 {
    panic("denominator must not be zero")
  }
  return n /d
}
```

And then you call it like this:

```go
div(userInput, 0.)
```

Even if we don't know `userInput`, we can see that this function call will panic in runtime.

The linter finds and reports such places using safe partial code execution and black magic.

## Installation

```bash
go install github.com/orsinium-labs/arguard@latest
```

## Usage

```bash
arguard ./...
```

Available flags:

* `-contracts.follow-imports`: set this flag to false to not extract contracts from the imported modules. In other words, contract (guard) violations will be reported only if the function with the contract and the function call are located in the same analyzed package. Useful for better **performance**.
* `-contracts.report-contracts`: emit a message for every detected contract. Useful for **debugging** to see if a contract was detected by the linter or not.
* `-arguard.report-errors`: set this flag to show failures during contract execution. By default, if arguard fails to execute a contract, it just moves on without reporting anything. Useful for **debugging** to see why a contract error wasn't reported.

## QnA

1. **How does it work?** There are two analyzers inside. The first one detects safe to execute contracts (guards) in the code. The second one detects calls to functions with knwon contracts, extracts statically known arguments, and executes contracts that can be executed using [yaegi](https://github.com/traefik/yaegi).
1. **What is a guard (contract)?** An if condition at the beginning of the function (only other contracts can go before it) with a safe to execute check and the body only returning an error or calling `panic`.
1. **How reliable are results?** If it reports an error, there is, most likely, an error. If it doesn't report an error, there still might be an error. It's a linter, not formal verifier.
1. **How stable is the project?** Static analysis in Go can be messy, especially when we also do partial code execution. The linter might fail, be wrong, or be not as smart as it potentially can be. Still, it's a static analyzer, not a production dependency, so it should be safe to use it on any project in any environment. Keep in mind, though, that there is still a partial code execution, so you probably shouldn't run it on untrusted code, just to be safe.
1. **Is there a [golangci-lint](https://golangci-lint.run/) integration?** Not yet but eventually will be. It's easy to integrate any [analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)-powered linter with golangci-lint, and arguard is analysis-powered. Stay tuned.
1. **Is there an IDE integration?** Not yet. When we have golangci-lint integration, [IDE integrations will come for free](https://golangci-lint.run/usage/integrations/).
1. **Is this a new idea?** The project implements one of the things that [deal](https://github.com/life4/deal), my Python library for [Design by Contract](https://en.wikipedia.org/wiki/Design_by_contract) can do. Deal itself is built upon wisdom of generations, see [this timeline](https://deal.readthedocs.io/basic/verification.html#background).
1. **Is it actively maintained?** The project, int the best traditions of UNIX-way, has a very small and clearly-defined scope. I might return to it time-to-time and bring new interesting ideas that I had during sleepless nights, but there is nothing to maintain daily. If it works today, it won't break tomorrow, thanks to the short list of dependencies and the [Go 1 compatibility promise](https://go.dev/doc/go1compat).
1. **What if I found a bug?** Fork the project, fix the bug, write some tests, and open a Pull Request. I usually merge and release any contributions within a day.
