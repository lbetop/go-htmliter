// SPDX-License-Identifier: MIT

package htmliter

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// copy from go.dev
const htmlString = `
<section class="Hero bluebg">
  <div class="Hero-gridContainer">
    <div class="Hero-blurb">
      <h1>Build simple, secure, scalable systems with Go</h1>
      <ul class="Hero-blurbList">
        <li>
          <svg width="12" height="10" viewBox="0 0 12 10" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10.8519 0.52594L3.89189 7.10404L1.14811 4.51081L0 5.59592L3.89189 9.27426L12 1.61105L10.8519 0.52594Z" fill="white" fill-opacity="0.87">
          </path></svg>
          An open-source programming language supported by Google
        </li>
        <li>
          <svg width="12" height="10" viewBox="0 0 12 10" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10.8519 0.52594L3.89189 7.10404L1.14811 4.51081L0 5.59592L3.89189 9.27426L12 1.61105L10.8519 0.52594Z" fill="white" fill-opacity="0.87">
          </path></svg>
          Easy to learn and great for teams
        </li>
        <li>
          <svg width="12" height="10" viewBox="0 0 12 10" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10.8519 0.52594L3.89189 7.10404L1.14811 4.51081L0 5.59592L3.89189 9.27426L12 1.61105L10.8519 0.52594Z" fill="white" fill-opacity="0.87">
          </path></svg>
          Built-in concurrency and a robust standard library
        </li>
        <li>
          <svg width="12" height="10" viewBox="0 0 12 10" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10.8519 0.52594L3.89189 7.10404L1.14811 4.51081L0 5.59592L3.89189 9.27426L12 1.61105L10.8519 0.52594Z" fill="white" fill-opacity="0.87">
          </path></svg>
          Large ecosystem of partners, communities, and tools
        </li>
      </ul>
    </div>
    <div class="Hero-actions">
      <div data-version="" class="js-latestGoVersion">
        <a class="Primary" href="/learn/" aria-label="Get Started" aria-describedby="getStarted-description" role="button">Get Started</a>
        <a class="Secondary js-downloadBtn" href="/dl" aria-label="Download" aria-describedby="download-description" role="button">Download</a>
        <div class="screen-reader-only" id="getStarted-description" hidden="">
          Opens a new window with the Get Started guide.
        </div>
        <div class="screen-reader-only" id="download-description" hidden="">
          Opens a new window to download Go.
        </div>
      </div>
      <div class="Hero-footnote">
        <p>
          Download packages for
          <a class="js-downloadWin" href="/dl/go1.27.3.windows-amd64.msi">Windows 64-bit</a>,
          <a class="js-downloadMac" href="/dl/go1.27.3.darwin-amd64.pkg">macOS</a>,
          <a class="js-downloadLinux" href="/dl/go1.27.3.linux-amd64.tar.gz">Linux</a>, and
          <a href="/dl/" aria-describedby="newwindow-description">more</a>
        </p>
        <p>
          The <code>go</code> command by default downloads and authenticates
          modules using the Go module mirror and Go checksum database run by
          Google. <a href="/dl" aria-describedby="newwindow-description">Learn more.</a>
        </p>
      </div>
    </div>
    <div class="screen-reader-only" id="newwindow-description" hidden="">
          Opens in new window.
    </div>
    <div class="Hero-gopher">
      <img class="Hero-gopherLadder" src="/images/gophers/ladder.svg" alt="Go Gopher climbing a ladder.">
    </div>
  </div>
</section>`

type condition func(*Token) bool
type assert func(*testing.T, *Tozer, *Token)

type testCase struct {
	cons [][]condition
	ass  assert
}

func dfs(t *testing.T, tozer *Tozer, token *Token, conditions []condition, i int, assert assert) bool {
	if i == len(conditions) {
		assert(t, tozer, token)
		return true
	}
	if !conditions[i](token) {
		return false
	}
	for newToken, err := range tozer.Children() {
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
			return true
		}
		canQuit := dfs(t, tozer, newToken, conditions, i+1, assert)
		if canQuit {
			return true
		}
	}
	return false
}

func TestGetTextContent(t *testing.T) {
	var r = strings.NewReader(htmlString)
	tozer := NewTozer(r)
	for token, err := range tozer.Children() {
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
		}
		if token.HasTag("div") && token.HasAttr("class", "Hero-blurb") {
			for token, err := range tozer.Children() {
				if err != nil && !errors.Is(err, io.EOF) {
					t.Error(err)
				}
				if token.HasTag("h1") {
					blurb := tozer.TextContent()
					want := "Build simple, secure, scalable systems with Go"
					if blurb != want {
						t.Errorf("result not match\nresult: %s\nwant:   %s", blurb, want)
					}
				}
			}
		}
	}
}

func TestGetAttrVal(t *testing.T) {
	testCases := []testCase{
		{
			cons: [][]condition{
				{
					func(token *Token) bool {
						return true
					},
					func(token *Token) bool {
						return token.HasTag("div") && token.HasAttr("class", "Hero-gridContainer")
					},
					func(token *Token) bool {
						return token.HasTag("div") && token.HasAttr("class", "Hero-gopher")
					},
					func(token *Token) bool {
						return token.HasTag("img") && token.HasAttr("class", "Hero-gopherLadder")
					},
				},
				{
					func(token *Token) bool {
						return true
					},
					func(token *Token) bool {
						return token.HasTag("img") && token.HasAttr("class", "Hero-gopherLadder")
					},
				},
			},
			ass: func(t *testing.T, tozer *Tozer, token *Token) {
				var img string
				for _, a := range token.Attr {
					if a.Key == "src" {
						img = a.Val
						break
					}
				}
				want := "/images/gophers/ladder.svg"
				if img != want {
					t.Errorf("result not match\nresult: %s\nwant:   %s", img, want)
				}
			},
		},
		{
			cons: [][]condition{
				{
					func(token *Token) bool {
						return true
					},
					func(token *Token) bool {
						return token.HasTag("div") && token.HasAttr("class", "Hero-actions")
					},
					func(token *Token) bool {
						return token.HasTag("img") && token.HasAttr("class", "Hero-gopherLadder")
					},
				},
			},
			ass: func(t *testing.T, tozer *Tozer, token *Token) {
				var img string
				for _, a := range token.Attr {
					if a.Key == "src" {
						img = a.Val
						break
					}
				}
				if len(img) > 0 {
					t.Error("don't expect to found the attribute")
				}
			},
		},
	}
	for _, cs := range testCases {
		for _, conBatch := range cs.cons {
			var r = strings.NewReader(htmlString)
			tozer := NewTozer(r)
			var token *Token
			dfs(t, tozer, token, conBatch, 0, cs.ass)
		}
	}
}

func TestGetTextContentRecursive(t *testing.T) {
	testCases := []testCase{
		{
			cons: [][]condition{
				{
					func(token *Token) bool {
						return true
					},
					func(token *Token) bool {
						return token.HasTag("div") && token.HasAttr("class", "Hero-blurb")
					},
					func(token *Token) bool {
						return token.HasTag("h1")
					},
				},
			},
			ass: func(t *testing.T, tozer *Tozer, token *Token) {
				blurb := tozer.TextContent()
				want := "Build simple, secure, scalable systems with Go"
				if blurb != want {
					t.Errorf("result not match\nresult: %s\nwant:   %s", blurb, want)
				}
			},
		},
	}
	for _, cs := range testCases {
		for _, con := range cs.cons {
			var r = strings.NewReader(htmlString)
			tozer := NewTozer(r)
			var token *Token
			dfs(t, tozer, token, con, 0, cs.ass)
		}
	}
}

func TestTagId(t *testing.T) {
	var r = strings.NewReader(htmlString)
	tozer := NewTozer(r)
	for token, err := range tozer.Children() {
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
		}
		if token.HasTag("div") && token.HasAttr("class", "Hero-blurb") {
			startId := tozer.id
			for token, err := range tozer.Children() {
				if err != nil && !errors.Is(err, io.EOF) {
					t.Error(err)
				}
				if token.HasTag("h1") {
					tozer.TextContent()
				}
			}
			endId := tozer.id
			if startId != endId+1 {
				t.Errorf("the matching mechanism is failed")
			}
		}
	}
}
