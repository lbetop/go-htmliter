# go-htmliter
go-htmliter simplifies [x/net/html.Token](https://pkg.go.dev/golang.org/x/net/html#Tokenizer) iteration by leveraging [Go iterators](https://pkg.go.dev/iter).

## Installation
```bash
go get github.com/lbetop/go-htmliter
```

## Example
```go
package main

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"

    "github.com/lbetop/go-htmliter"
)

func main() {
    res, err := http.Get("https://go.dev")
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
    }

    tozer := htmliter.NewTozer(res.Body)
    for token, err := range tozer.Children() {
        if err != nil && !errors.Is(err, io.EOF) {
            log.Fatal(err)
        }
        if token.HasTag("div") && token.HasAttr("class", "Hero-blurb") {
            for token, err := range tozer.Children() {
                if err != nil && !errors.Is(err, io.EOF) {
                    log.Fatal(err)
                }
                if token.HasTag("h1") {
                    blurb := tozer.TextContent()
                    fmt.Println(blurb)
                }
            }
        }
    }
}
```
