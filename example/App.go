package main

import (
	"fmt"
	"strconv"
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
	age, _ := o.UseState(42)
	_bool, _ := o.UseState(false)
	_flt, _ := o.UseState(0.3)
	p, setP := o.UseState(Person{})

	handleSubmit := func(_ js.Value) {
		age_i, _ := strconv.Atoi((*age).(string))
		setP(Person{(*name).(string), age_i, (*_bool).(bool), (*_flt).(float64)})
		fmt.Println((*p).(Person))
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
			),
		),
	)
}
