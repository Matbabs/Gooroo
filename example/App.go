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

	o.Css("App.css")

	name, _ := o.UseState("Paul")
	age, _ := o.UseState("42")
	p, setP := o.UseState(Person{(*name).(string), (*age).(string)})

	arr := []string{"Cat", "Dog", "Bird"}

	handleSubmit := func(_ js.Value) {
		setP(Person{(*name).(string), (*age).(string)})
	}

	return o.Div(o.Style("margin: auto; padding: 100px; width: 800px"),
		o.H1(
			"Tuto Gooroo web front Go ! v0.0.8",
			o.ClassName("title"),
		),
		o.H2("Form"),
		o.Div(o.GridLayout(3, 0, "20px"),
			o.Span("Name"),
			o.Span("Age"),
			o.Span("42"),
			o.Input(o.OnChange(name)),
			o.Input(o.OnChange(age), o.Type("number")),
			o.Button("Click", o.OnClick(handleSubmit), o.Title("")),
		),
		o.H2("Result"),
		o.If(*p != nil,
			o.P((*p).(Person).name),
			o.P((*p).(Person).age),
		),
		o.H2("Animals"),
		o.Div(
			o.For(arr, func(i int) o.DomComponent {
				return o.P(arr[i])
			}),
		),
	)
}
