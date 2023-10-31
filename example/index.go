package main

import (
	o "github.com/Matbabs/Gooroo"
	"github.com/Matbabs/Gooroo/example/components"
)

func main() {
	o.Render(func() {
		o.Html(
			components.App(),
		)
	})
}
