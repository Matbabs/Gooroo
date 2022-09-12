package main

import (
	"fmt"
	"syscall/js"

	o "github.com/Matbabs/Gooroo"
)

type Person struct {
	name string
	age  int
	has  bool
	flt  float64
}

func App() o.DomComponent {

	o.Css("")

	name, _ := o.UseState("")
	age, _ := o.UseState("42")
	_bool, _ := o.UseState("FALSE")
	_flt, _ := o.UseState("0.3")
	p, setP := o.UseState(Person{})

	handleSubmit := func(e js.Value) {
		setP(Person{o.AnyStr(*name), o.AnyInt(*age), o.AnyBol(*_bool), o.AnyFlt(*_flt)})
	}

	handleMemo := o.UseMemo(func() any {
		return 1 + 2
	}, name)

	fmt.Println(handleMemo)

	return o.Div(o.Style("margin: auto; padding: 100px; width: 800px"),
		o.H1("Tuto BreizhC@mp web front Go ! v0.0.2"),
		o.H2("Form"),
		o.Div(o.GridLayout(3, 0, "20px"),
			o.Span("Name"),
			o.Span("Age"),
			o.Span(nil),
			o.Input(o.OnChange(name)),
			o.Input(o.OnChange(age), o.Type("number")),
			o.Button("Click", o.OnClick(handleSubmit), o.Title(666)),
			o.If(*age != nil,
				o.P(*p),
			),
		),
	)
}
