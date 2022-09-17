// GOOS=js GOARCH=wasm go test -cover -o example/main.wasm

package gooroo

import (
	"fmt"
	"strings"
	"syscall/js"
	"testing"

	"github.com/Matbabs/Gooroo/dom"
)

type test struct {
	name     string
	function func(t *testing.T)
}

var head js.Value
var body js.Value

var tests = []test{
	{
		"Css",
		func(t *testing.T) {
			path := "./master.css"
			Css(path)
			href := head.Get(dom.JS_CHILDREN).Get("4").Get(dom.JS_HREF).String()
			if !strings.Contains(href, "master.css") {
				t.Error("master.css not exist in href")
			}
		},
	},
	{
		"Html void",
		func(t *testing.T) {
			Html()
			length := body.Get(dom.JS_CHILDREN).Get(dom.JS_LENGTH).String()
			if !strings.Contains(length, "number: 1") {
				t.Error("Body has different number of child than expected")
			}
		},
	},
	{
		"Html children",
		func(t *testing.T) {
			title := "Test title !"
			paragraph := "Paragraph test"
			Html(
				Div(
					H1(title),
					P(paragraph),
				),
			)
			div := body.Get(dom.JS_CHILDREN).Get("1").Get(dom.JS_CHILDREN).Get("0")
			h1 := div.Get(dom.JS_CHILDREN).Get("0").Get(dom.JS_INNER_HTML)
			p := div.Get(dom.JS_CHILDREN).Get("1").Get(dom.JS_INNER_HTML)
			if title != h1.String() || paragraph != p.String() {
				t.Error("Innerhtml does not contain expected value")
			}
		},
	},
	{
		"sanitizeHtml",
		func(t *testing.T) {
			str := "<script>TEST</script>"
			sanitizeHtml(&str)
			if strings.Contains(str, "<script>") {
				t.Error("The tag <script> is present")
			}
		},
	},
	{
		"clearContext",
		func(t *testing.T) {
			clearContext()
			if document.Get(dom.HTML_BODY).Get(dom.JS_INNER_HTML).String() != "" {
				t.Error("Innerhtml is not equal to void")
			}
		},
	},
}

func Test_All(t *testing.T) {

	if document.IsUndefined() {
		t.Skip("Not a browser environment. Skipping.")
	}

	head = document.Get(dom.HTML_HEAD)
	body = document.Get(dom.HTML_BODY)

	for _, test := range tests {
		fmt.Println(fmt.Sprintf("Test: %s", test.name))
		t.Run(test.name, test.function)
	}

}
