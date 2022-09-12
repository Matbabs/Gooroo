package main

import (
	o "github.com/Matbabs/Gooroo"
)

func main() {
	o.Render(func() {
		o.Html(
			App(),
		)
	})
}
