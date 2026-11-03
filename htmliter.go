// SPDX-License-Identifier: MIT

// Package htmliter provides a way to iterate over an [x/net/html.Tokenizer].
// It assumes the provided HTML input is well-formed and valid.
// Invalid or malformed HTML structures may cause unpredictable results.
//
// For more information, read [x/net/html].
//
// [x/net/html]: https://pkg.go.dev/golang.org/x/net/html
// [x/net/html.Tokenizer]: https://pkg.go.dev/golang.org/x/net/html#Tokenizer
package htmliter

import (
	"io"
	"iter"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Tozer is a wrapper around [x/net/html.Tokenizer].
// It implements a tag-indexing mechanism to trace depth hierarchy.
//
// [x/net/html.Tokenizer]: https://pkg.go.dev/golang.org/x/net/html#Tokenizer
type Tozer struct {
	tokenizer html.Tokenizer
	id        int
}

// TokenType represents the type of an HTML token.
type TokenType html.TokenType

// HTML token types mirrored from [x/net/html.TokenType].
//
// [x/net/html.TokenType]: https://pkg.go.dev/golang.org/x/net/html#TokenType
const (
	ErrorToken          = TokenType(html.ErrorToken)
	TextToken           = TokenType(html.TextToken)
	StartTagToken       = TokenType(html.StartTagToken)
	EndTagToken         = TokenType(html.EndTagToken)
	SelfClosingTagToken = TokenType(html.SelfClosingTagToken)
	CommentToken        = TokenType(html.CommentToken)
	DoctypeToken        = TokenType(html.DoctypeToken)
)

// Attribute represents a key-value pair belonging to an HTML tag.
type Attribute struct {
	Key, Val string
}

// Token encapsulates the HTML token's data, type, and its associated attributes.
type Token struct {
	Data string
	Type TokenType
	Attr []Attribute
}

// voidElements defines all HTML elements that are standalone and do not require a closing tag.
var voidElements = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "embed": {},
	"hr": {}, "img": {}, "input": {}, "keygen": {}, "link": {},
	"meta": {}, "param": {}, "source": {}, "track": {}, "wbr": {},
}

// Children returns a single-use iterator over the immediate child tokens inside the current token block.
func (tozer *Tozer) Children() iter.Seq2[*Token, error] {
	return func(yield func(*Token, error) bool) {
		id := tozer.id - 1
		for tozer.id > id {
			token := tozer.next()
			switch token.Type {
			case ErrorToken:
				yield(token, tozer.tokenizer.Err())
				return
			case StartTagToken, SelfClosingTagToken:
				if !yield(token, nil) {
					return
				}
			}
		}
	}
}

// TextContent returns the concatenated text content inside the current node block.
// The output is like [Node.textContent].
// Since this method consumes the underlying tokenizer stream to gather text,
// subsequent calls will not produce the same output.
// You must assign the result to a variable if you need to reuse it.
//
// [Node.textContent]: https://developer.mozilla.org/en-US/docs/Web/API/Node/textContent
func (tozer *Tozer) TextContent() string {
	var res strings.Builder
	id := tozer.id - 1
	for tozer.id > id {
		token := tozer.next()
		switch token.Type {
		case ErrorToken:
			return res.String()
		case TextToken:
			res.WriteString(token.Data)
		}
	}
	return res.String()
}

// SetMaxBuf sets a limit on the maximum data buffered during tokenization to prevent excessive memory usage.
func (tozer *Tozer) SetMaxBuf(n int) {
	tozer.tokenizer.SetMaxBuf(n)
}

// next wraps the internal tokenizer.Next call, updates the depth id, and returns a parsed Token.
func (tozer *Tozer) next() *Token {
	token := Token{Type: TokenType(tozer.tokenizer.Next())}
	switch token.Type {
	case StartTagToken, SelfClosingTagToken, EndTagToken:
		bytes, remaining := tozer.tokenizer.TagName()
		token.Data = atom.String(bytes)
		for remaining {
			var key, value []byte
			key, value, remaining = tozer.tokenizer.TagAttr()
			token.Attr = append(token.Attr, Attribute{string(key), string(value)})
		}
		if _, exist := voidElements[token.Data]; !exist {
			tozer.updateId(token.Type)
		}
	case TextToken:
		token.Data = string(tozer.tokenizer.Text())
	}
	return &token
}

// updateId increments or decrements the internal depth tracer based on the tag type.
// It assumes the underlying HTML is valid.
func (tozer *Tozer) updateId(tokenType TokenType) {
	switch tokenType {
	case StartTagToken:
		tozer.id++
	case EndTagToken:
		tozer.id--
	}
}

// HasTag checks if the token's tagname matches the specified tag string.
func (token *Token) HasTag(tag string) bool {
	return token.Data == tag
}

// HasAttr checks if the token contains an attribute matching the exact key and value.
func (token *Token) HasAttr(key, val string) bool {
	for _, a := range token.Attr {
		if a.Key == key {
			return a.Val == val
		}
	}
	return false
}

// HasPrefixAttr checks if the token contains an attribute with the specified key whose value starts with the given prefix.
func (token *Token) HasPrefixAttr(key, prefix string) bool {
	for _, a := range token.Attr {
		if a.Key == key {
			return strings.HasPrefix(a.Val, prefix)
		}
	}
	return false
}

// NewTozer returns a new Tozer instance for the given Reader.
func NewTozer(r io.Reader) *Tozer {
	return &Tozer{*html.NewTokenizer(r), 0}
}
