// The Gooroo package gathers a set of cumulative functions allowing you
// to create web applications on the Frontend side.
// To do this purpose, it implements DOM manipulation features based on syscall/js
// and webassembly.
// Its objective is to explore the possibilities of a modern, lightweight and
// javascript independent web library.
package gooroo

import (
	"fmt"
	"runtime"
	"strings"
	"syscall/js"

	"github.com/Matbabs/Gooroo/dom"
	"github.com/Matbabs/Gooroo/utils"
)

// DomComponent represents an element of the DOM. This element can be a tag, an attribute,
// a layout or even a binding. Most DomComponents can be nested within each other thanks to
// variadic parameters.
type DomComponent func() string

// DomBinding is a structure that allows to retain the link between an event, a callback function
// and a potential value to update when the event is triggered.
// All binding is reapplied during rendering.
type domBinding struct {
	event    string
	callback js.Func
	value    *any
}

// DomStore allows to keep the state of change of a value in the store.
type domStore struct {
	value      any
	hasChanged bool
}

// Represents the global variable "document" of a website, useful for DOM manipulation and some
// JavaScript interraction.
var document js.Value = js.Global().Get(dom.HTML_DOCUMENT)

// List of paths to add CSS style sheets already imported into the website.
var stylesheets = []string{}

// Communication channel that generates a new rendering for each message sent within it.
var state = make(chan bool)

// List of DomBindings registered for the application rendering.
var bindings = make(map[string][]domBinding)

// Store of local variables recorded in the application state.
var store = make(map[string]*domStore)

// Store of memoized variables.
var storeMemo = make(map[string]any)

// Store of memoized functions.
var storeCallback = make(map[string]*func(...any) any)

// Manipulate DOM

// Hangs a CSS file in the <head> content of the website.
func Css(filepath string) {
	if !utils.Contains(stylesheets, filepath) {
		stylesheets = append(stylesheets, filepath)
		elem := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_LINK)
		document.Get(dom.HTML_HEAD).Call(dom.JS_APPEND_CHILD, elem)
		elem.Set(dom.JS_REL, dom.HTML_STYLESHEET)
		elem.Set(dom.JS_HREF, filepath)
	}
}

// Triggers a rendering of the DOM, of all the DomComponents declared in parameters.
func Html(domComponents ...DomComponent) {
	for i := range domComponents {
		elem := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_DIV)
		document.Get(dom.HTML_BODY).Call(dom.JS_APPEND_CHILD, elem)
		elem.Set(dom.JS_INNER_HTML, domComponents[i]())
	}
	setBindings()
}

// Clean the content of a string to prevent injections related to innerHtml attributes.
func sanitizeHtml(htmlStr *string) {
	tmp := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_DIV)
	tmp.Set(dom.JS_INNER_HTML, *htmlStr)
	*htmlStr = tmp.Get(dom.JS_TEXT_CONTENT).String()
}

// Removes the whole rendering from the DOM, including the DomComponents passed as parameters
// in the Html() function.
func clearContext() {
	document.Get(dom.HTML_BODY).Set(dom.JS_INNER_HTML, "")
}

// Create a functional DomBinding set on its parameters.
func generateBinding(event string, value *any, callbacks ...func(js.Value)) domBinding {
	return domBinding{
		event,
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				if event == dom.JS_EVENT_KEYUP || event == dom.JS_EVENT_CHANGE {
					*value = args[0].Get(dom.JS_TARGET).Get(dom.JS_VALUE).String()
				}
				for i := range callbacks {
					callbacks[i](args[0])
				}
				return nil
			},
		),
		value,
	}
}

// Applies all the bindings to the DOM elements concerned.
func setBindings() {
	for id := range bindings {
		elem := document.Call(dom.JS_GET_ELEMENT_BY_ID, id)
		for i := range bindings[id] {
			elem.Call(dom.JS_ADD_EVENT_LISTENER, bindings[id][i].event, bindings[id][i].callback)
			if bindings[id][i].event == dom.JS_EVENT_CHANGE {
				elem.Set(dom.JS_VALUE, *(bindings[id][i].value))
			}
		}
	}
}

// Deletes all the DomBindings stored locally.
func unsetBindings() {
	bindings = make(map[string][]domBinding)
}

// Checks if one or more variables in the store have been changed.
func detectHasChanged(variables ...*any) bool {
	for i := range variables {
		for key := range store {
			if variables[i] == &store[key].value && store[key].hasChanged {
				return true
			}
		}
	}
	return false
}

// Reset to zero of all the changes of each variable stored in the store.
func clearHasChange() {
	for k := range store {
		if store[k].hasChanged {
			store[k].hasChanged = false
		}
	}
}

// Starts the library's renderer. Allows to re-trigger the renderings when the
// state changes (with a UseSate variable for example), through the state channel.
// Must take a lambda function func() containing the call to Html() as parameter
// to execute a rendering context.
func Render(context func()) {
	updateState()
	for {
		<-state
		clearContext()
		unsetBindings()
		context()
		clearHasChange()
	}
}

// Called to trigger in parallel a message sending in the chan state and consequently
// request the new rendering of the application.
func updateState() {
	go func() {
		state <- true
	}()
}

// Returns a stateful value, and a function to update it.
// During the initial render, the returned state (state) is the same as the value
// passed as the first argument (initialState).
// The setState function is used to update the state. It accepts a new state value
// and enqueues a re-render of the DOM.
func UseState(initialValue any) (actualValue *any, f func(setterValue any)) {
	_, file, no, _ := runtime.Caller(1)
	key := utils.CallerToKey(file, no)
	utils.MapInit(key, store, &domStore{initialValue, false})
	return &store[key].value, func(setVal any) {
		store[key].value = setVal
		store[key].hasChanged = true
		updateState()
	}
}

// Accepts a function that contains imperative, possibly effectful code.
// The default behavior for effects is to fire the effect after every completed render.
// That way an effect is always recreated if one of its variables changes (in variadics params).
func UseEffect(callback func(), variables ...*any) {
	if len(variables) == 0 || detectHasChanged(variables...) {
		callback()
	}
}

// Pass an inline callback and an array of dependencies. useCallback will return a memoized
// version of the callback that only changes if one of the dependencies has changed.
func UseCallback(callback func(...any) any, variables ...*any) *func(...any) any {
	_, file, no, _ := runtime.Caller(1)
	key := utils.CallerToKey(file, no)
	utils.MapInit(key, storeCallback, &callback)
	if len(variables) == 0 || detectHasChanged(variables...) {
		storeCallback[key] = &callback
	}
	return storeCallback[key]
}

// Pass a “create” function and an array of dependencies. useMemo will only recompute the memoized
// value when one of the dependencies has changed. This optimization helps to avoid expensive
// calculations on every render.
func UseMemo(callback func() any, variables ...*any) any {
	_, file, no, _ := runtime.Caller(1)
	key := utils.CallerToKey(file, no)
	utils.MapInitCallback(key, storeMemo, callback)
	if len(variables) == 0 || detectHasChanged(variables...) {
		storeMemo[key] = callback()
	}
	return storeMemo[key]
}

// Generate code for the DOM

// Fires again if one of the insiders has a dom.ELEMENT_PARAM anchor, to hook
// the parameter as an attribute of the HTML element passed to the DOM.
func insertDomComponentParams(opener string, insiders ...DomComponent) (string, []DomComponent) {
	var insidersWithoutParam []DomComponent
	for _, insider := range insiders {
		if strings.Contains(insider(), dom.ELEMENT_PARAM) {
			split := (strings.Split(opener, dom.ELEMENT_PARAM_SPLIT))
			opener = fmt.Sprintf("%s %s%s%s", split[0], strings.Split(insider(), dom.ELEMENT_PARAM)[1], dom.ELEMENT_PARAM_SPLIT, split[1])
		} else {
			insidersWithoutParam = append(insidersWithoutParam, insider)
		}
	}
	return opener, insidersWithoutParam
}

// Returns the html rendering of the DomComponent reproduced recursively with all its DomComponents insiders
func htmlDomComponent(opener string, closer string, insiders ...DomComponent) DomComponent {
	if len(insiders) > 0 {
		htmlStr, insiders := insertDomComponentParams(opener, insiders...)
		for _, insider := range insiders {
			htmlStr += insider()
		}
		htmlStr += closer
		return func() string { return htmlStr }
	}
	return func() string { return fmt.Sprintf("%s%s", opener, closer) }
}

// Same function as htmlDomComponent() but only if the condition in parameter is valid.
func If(condition bool, insiders ...DomComponent) DomComponent {
	if condition {
		htmlStr := ""
		for _, insider := range insiders {
			htmlStr += insider()
		}
		return func() string { return htmlStr }
	}
	return func() string { return "" }
}

// Same operation as htmlDomComponent() but applies the function passed in parameter for the
// whole array. The "key" element is used to make the link with the elements within the function.
func For[T string | int | int32 | int64 | float32 | float64 | bool | any](elements []T, keyDomComponent func(i int) DomComponent) DomComponent {
	if len(elements) > 0 {
		htmlStr := ""
		for i := range elements {
			htmlStr += keyDomComponent(i)()
		}
		return func() string { return htmlStr }
	}
	return func() string { return "" }
}

// DomComponents

// Declare an html element with the <div> tag.
func Div(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_DIV_OPENER, dom.HTML_DIV_CLOSER, insiders...)
}

// Declare an html element with the <p> tag.
func P[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_P_OPENER, textStr), dom.HTML_P_CLOSER, insiders...)
}

// Declare an html element with the <span> tag.
func Span[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_SPAN_OPENER, textStr), dom.HTML_SPAN_CLOSER, insiders...)
}

// Declare an html element with the <ul> tag.
func Ul(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_UL_OPENER, dom.HTML_UL_CLOSER, insiders...)
}

// Declare an html element with the <li> tag.
func Li(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_LI_OPENER, dom.HTML_LI_CLOSER, insiders...)
}

// Declare an html element with the <table> tag.
func Table(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TABLE_OPENER, dom.HTML_TABLE_CLOSER, insiders...)
}

// Declare an html element with the <tr> tag.
func Tr(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TR_OPENER, dom.HTML_TR_CLOSER, insiders...)
}

// Declare an html element with the <th> tag.
func Th(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TH_OPENER, dom.HTML_TH_CLOSER, insiders...)
}

// Declare an html element with the <td> tag.
func Td(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TD_OPENER, dom.HTML_TD_CLOSER, insiders...)
}

// Declare an html element with the <h1> tag.
func H1[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H1_OPENER, textStr), dom.HTML_H1_CLOSER, insiders...)
}

// Declare an html element with the <h2> tag.
func H2[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H2_OPENER, textStr), dom.HTML_H2_CLOSER, insiders...)
}

// Declare an html element with the <h3> tag.
func H3[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H3_OPENER, textStr), dom.HTML_H3_CLOSER, insiders...)
}

// Declare an html element with the <h4> tag.
func H4[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H4_OPENER, textStr), dom.HTML_H4_CLOSER, insiders...)
}

// Declare an html element with the <a> tag.
func A[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_A_OPENER, textStr), dom.HTML_A_CLOSER, insiders...)
}

// Declare an html element with the <form> tag.
func Form(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_FORM_OPENER, dom.HTML_FORM_CLOSER, insiders...)
}

// Declare an html element with the <textarea> tag.
func TextArea(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TEXTAREA_OPENER, dom.HTML_TEXTAREA_CLOSER, insiders...)
}

// Declare an html element with the <select> tag.
func Select(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_SELECT_OPENER, dom.HTML_SELECT_CLOSER, insiders...)
}

// Declare an html element with the <option> tag.
func Option[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_OPTION_OPENER, textStr), dom.HTML_OPTION_CLOSER, insiders...)
}

// Declare an html element with the <input> tag.
func Input(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_INPUT_OPENER, "", insiders...)
}

// Declare an html element with the <button> tag.
func Button[T string | int | int32 | int64 | float32 | float64 | bool](text T, insiders ...DomComponent) DomComponent {
	textStr := utils.AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_BUTTON_OPENER, textStr), dom.HTML_BUTTON_CLOSER, insiders...)
}

// Declare an html element with the <img> tag.
func Img(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_IMG_OPENER, "", insiders...)
}

// Declare an html element with the <br> tag.
func Br() DomComponent {
	return htmlDomComponent(dom.HTML_BR_OPENER, "")
}

// DomComponentsParams

// Declare an attribute of an html element with the value 'class='
func ClassName(className string) DomComponent {
	sanitizeHtml(&className)
	return func() string {
		return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_CLASSNAME, className)
	}
}

// Declare an attribute of an html element with the value 'style='
func Style(style string) DomComponent {
	sanitizeHtml(&style)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, style) }
}

// Declare an attribute of an html element with the value 'href='
func Href(href string) DomComponent {
	sanitizeHtml(&href)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_HREF, href) }
}

// Declare an attribute of an html element with the value 'src='
func Src(src string) DomComponent {
	sanitizeHtml(&src)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_SRC, src) }
}

// Declare an attribute of an html element with the value 'value='
func Value(value string) DomComponent {
	sanitizeHtml(&value)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_VALUE, value) }
}

// Declare an attribute of an html element with the value 'id='
func Id(id string) DomComponent {
	sanitizeHtml(&id)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TYPE, id) }
}

// Declare an attribute of an html element with the value 'type='
func Type(_type string) DomComponent {
	sanitizeHtml(&_type)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TYPE, _type) }
}

// Declare an attribute of an html element with the value 'placeholder='
func Placeholder(placeholder string) DomComponent {
	sanitizeHtml(&placeholder)
	return func() string {
		return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_PLACEHOLDER, placeholder)
	}
}

// Declare an attribute of an html element with the value 'title='
func Title(title string) DomComponent {
	sanitizeHtml(&title)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TITLE, title) }
}

// DomComponentsParamsStructure

// Declare une configuration CSS dans l'attribut d'un element html avec la valeur 'style=',
// de manière a paramettrer un 'display: flex'
func FlexLayout[T string | int](flow string, justify string, align string, gap T) DomComponent {
	gapStr := utils.AnyStr(gap)
	sanitizeHtml(&flow)
	sanitizeHtml(&justify)
	sanitizeHtml(&align)
	sanitizeHtml(&gapStr)
	layout := fmt.Sprintf("%s %s;%s %s;%s %s;%s %s;%s %s", dom.CSS_PARAM_DISPLAY, dom.CSS_PARAM_DISPLAY_FLEX,
		dom.CSS_PARAM_FLOW, flow, dom.CSS_PARAM_JUSTIFY, justify, dom.CSS_PARAM_ALIGN, align, dom.CSS_PARAM_GAP, gapStr)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, layout) }
}

// Declare une configuration CSS dans l'attribut d'un element html avec la valeur 'style=',
// de manière a paramettrer un 'display: grid'
func GridLayout[T string | int](columns T, rows T, gap string) DomComponent {
	columnsStr := utils.AnyStr(columns)
	rowsStr := utils.AnyStr(rows)
	sanitizeHtml(&columnsStr)
	sanitizeHtml(&rowsStr)
	sanitizeHtml(&gap)
	layout := fmt.Sprintf("%s %s;%s %s%s%s;%s %s%s%s;%s %s", dom.CSS_PARAM_DISPLAY, dom.CSS_PARAM_DISPLAY_GRID,
		dom.CSS_PARAM_GRID_COLUMNS, dom.CSS_PARAM_GRID_REPEAT_OPENER, columnsStr, dom.CSS_PARAM_GRID_REPEAT_CLOSER,
		dom.CSS_PARAM_GRID_ROWS, dom.CSS_PARAM_GRID_REPEAT_OPENER, rowsStr, dom.CSS_PARAM_GRID_REPEAT_CLOSER, dom.CSS_PARAM_GAP, gap)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, layout) }
}

// DomComponentsParamsBinding

// Declare a binding on the event 'click' on the attached element to trigger
// the function passed in parameter.
func OnClick(callbacks ...func(js.Value)) DomComponent {
	id := utils.GenerateShortId()
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_CLICK, nil, callbacks...))
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_ID, id) }
}

// Declare a binding on the event 'change' on the attached element to trigger
// the function passed in parameter.
func OnChange(value *any, callbacks ...func(js.Value)) DomComponent {
	id := utils.GenerateShortId()
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_CHANGE, value, callbacks...))
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_KEYUP, value, callbacks...))
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_ID, id) }
}
