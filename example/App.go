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

	o.UseEffect(func() {
		fmt.Println("Has changed")
	}, p)

	o.UseEffect(func() {
		fmt.Println("Has changed 2")
	}, name)

	handleSubmit := func(e js.Value) {
		fmt.Println(*_bool)
		setP(Person{o.AnyStr(*name), o.AnyInt(*age), o.AnyBol(*_bool), o.AnyFlt(*_flt)})
	}

	handleCall := o.UseCallback(func(a ...any) any {
		fmt.Println("t")
		return nil
	}, p)

	fmt.Println(handleCall)

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
