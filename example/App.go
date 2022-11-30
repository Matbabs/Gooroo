package main

import (
	"syscall/js"

	o "github.com/Matbabs/Gooroo"
)

type Person struct {
	name string
	age  string
}

func App() o.DomComponent {

	o.Css("")

	name, _ := o.UseState("")
	age, _ := o.UseState("42")
	p, setP := o.UseState(Person{})

	handleSubmit := func(_ js.Value) {
		setP(Person{(*name).(string), (*age).(string)})
	}

	return o.Div(o.Style("margin: auto; padding: 100px; width: 800px"),
		o.H1("Tuto BreizhC@mp web front Go ! v0.0.2"),
		o.H2("Form"),
		o.Div(o.GridLayout(3, 0, "20px"),
			o.Span("Name"),
			o.Span("Age"),
			o.Span("42"),
			o.Input(o.OnChange(name)),
			o.Input(o.OnChange(age), o.Type("number")),
			o.Button("Click", o.OnClick(handleSubmit), o.Title("")),
			o.If(*age != nil,
				o.P((*p).(Person).name),
				o.P((*p).(Person).age),
			),
		),
	)
}
