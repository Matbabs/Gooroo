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
	"strconv"
	"strings"
	"syscall/js"

	"github.com/Matbabs/Gooroo/dom"
	"github.com/Matbabs/Gooroo/utils"
)

type DomComponent func() string
type DomBinding struct {
	event    string
	callback js.Func
	value    *any
}
type DomStore struct {
	value      any
	hasChanged bool
}

var document js.Value = js.Global().Get(dom.HTML_DOCUMENT)
var stylesheets = []string{}
var state = make(chan bool)
var bindings = make(map[string][]DomBinding)
var store = make(map[string]*DomStore)
var storeMemo = make(map[string]any)
var storeCallback = make(map[string]*func(...any) any)

// Manipulate DOM

func Css(filepath string) {
	if !utils.Contains(stylesheets, filepath) {
		stylesheets = append(stylesheets, filepath)
		elem := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_LINK)
		document.Get(dom.HTML_HEAD).Call(dom.JS_APPEND_CHILD, elem)
		elem.Set(dom.JS_REL, dom.HTML_STYLESHEET)
		elem.Set(dom.JS_HREF, filepath)
	}
}

func Html(domComponents ...DomComponent) {
	for i := range domComponents {
		elem := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_DIV)
		document.Get(dom.HTML_BODY).Call(dom.JS_APPEND_CHILD, elem)
		elem.Set(dom.JS_INNER_HTML, domComponents[i]())
	}
	setBindings()
}

func sanitizeHtml(htmlStr *string) {
	tmp := document.Call(dom.JS_CREATE_ELEMENT, dom.HTML_DIV)
	tmp.Set(dom.JS_INNER_HTML, *htmlStr)
	*htmlStr = tmp.Get(dom.JS_TEXT_CONTENT).String()
}

func clearContext() {
	document.Get(dom.HTML_BODY).Set(dom.JS_INNER_HTML, "")
}

func generateBinding(event string, value *any, callbacks ...func(js.Value)) DomBinding {
	return DomBinding{
		event,
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				if event == dom.JS_EVENT_CHANGE {
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

func unsetBindings() {
	bindings = make(map[string][]DomBinding)
}

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

func clearHasChange() {
	for k := range store {
		if store[k].hasChanged {
			store[k].hasChanged = false
		}
	}
}

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

func updateState() {
	go func() {
		state <- true
	}()
}

func UseState(initialValue any) (actualValue *any, f func(setterValue any)) {
	_, file, no, _ := runtime.Caller(1)
	key := utils.CallerToKey(file, no)
	utils.MapInit(key, store, &DomStore{initialValue, false})
	return &store[key].value, func(setVal any) {
		store[key].value = setVal
		store[key].hasChanged = true
		updateState()
	}
}

func UseEffect(callback func(), variables ...*any) {
	if len(variables) == 0 || detectHasChanged(variables...) {
		callback()
	}
}

func UseCallback(callback func(...any) any, variables ...*any) *func(...any) any {
	_, file, no, _ := runtime.Caller(1)
	key := utils.CallerToKey(file, no)
	utils.MapInit(key, storeCallback, &callback)
	if len(variables) == 0 || detectHasChanged(variables...) {
		storeCallback[key] = &callback
	}
	return storeCallback[key]
}

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

func For(elements []any, keyDomComponent func(i int) DomComponent) DomComponent {
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

func Div(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_DIV_OPENER, dom.HTML_DIV_CLOSER, insiders...)
}

func P[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_P_OPENER, textStr), dom.HTML_P_CLOSER, insiders...)
}

func Span[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_SPAN_OPENER, textStr), dom.HTML_SPAN_CLOSER, insiders...)
}

func Ul(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_UL_OPENER, dom.HTML_UL_CLOSER, insiders...)
}

func Li(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_LI_OPENER, dom.HTML_LI_CLOSER, insiders...)
}

func Table(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TABLE_OPENER, dom.HTML_TABLE_CLOSER, insiders...)
}

func Tr(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TR_OPENER, dom.HTML_TR_CLOSER, insiders...)
}

func Th(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TH_OPENER, dom.HTML_TH_CLOSER, insiders...)
}

func Td(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TD_OPENER, dom.HTML_TD_CLOSER, insiders...)
}

func H1[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H1_OPENER, textStr), dom.HTML_H1_CLOSER, insiders...)
}

func H2[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H2_OPENER, textStr), dom.HTML_H2_CLOSER, insiders...)
}

func H3[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H3_OPENER, textStr), dom.HTML_H3_CLOSER, insiders...)
}

func H4[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_H4_OPENER, textStr), dom.HTML_H4_CLOSER, insiders...)
}

func A[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_A_OPENER, textStr), dom.HTML_A_CLOSER, insiders...)
}

func Form(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_FORM_OPENER, dom.HTML_FORM_CLOSER, insiders...)
}

func TextArea(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_TEXTAREA_OPENER, dom.HTML_TEXTAREA_CLOSER, insiders...)
}

func Select(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_SELECT_OPENER, dom.HTML_SELECT_CLOSER, insiders...)
}

func Option[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_OPTION_OPENER, textStr), dom.HTML_OPTION_CLOSER, insiders...)
}

func Input(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_INPUT_OPENER, "", insiders...)
}

func Button[T string | int | int32 | int64 | float32 | float64](text T, insiders ...DomComponent) DomComponent {
	textStr := AnyStr(text)
	sanitizeHtml(&textStr)
	return htmlDomComponent(fmt.Sprintf("%s%s", dom.HTML_BUTTON_OPENER, textStr), dom.HTML_BUTTON_CLOSER, insiders...)
}

func Img(insiders ...DomComponent) DomComponent {
	return htmlDomComponent(dom.HTML_IMG_OPENER, "", insiders...)
}

func Br() DomComponent {
	return htmlDomComponent(dom.HTML_BR_OPENER, "")
}

// DomComponentsParams

func ClassName(className string) DomComponent {
	sanitizeHtml(&className)
	return func() string {
		return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_CLASSNAME, className)
	}
}

func Style(style string) DomComponent {
	sanitizeHtml(&style)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, style) }
}

func Href(href string) DomComponent {
	sanitizeHtml(&href)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_HREF, href) }
}

func Src(src string) DomComponent {
	sanitizeHtml(&src)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_SRC, src) }
}

func Value(value string) DomComponent {
	sanitizeHtml(&value)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_VALUE, value) }
}

func Id(id string) DomComponent {
	sanitizeHtml(&id)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TYPE, id) }
}

func Type(_type string) DomComponent {
	sanitizeHtml(&_type)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TYPE, _type) }
}

func Placeholder(placeholder string) DomComponent {
	sanitizeHtml(&placeholder)
	return func() string {
		return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_PLACEHOLDER, placeholder)
	}
}

func Title(title string) DomComponent {
	sanitizeHtml(&title)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_TITLE, title) }
}

// DomComponentsParamsStructure

func FlexLayout[T string | int | int32 | int64 | float32 | float64](flow string, justify string, align string, gap T) DomComponent {
	gapStr := AnyStr(gap)
	sanitizeHtml(&flow)
	sanitizeHtml(&justify)
	sanitizeHtml(&align)
	sanitizeHtml(&gapStr)
	layout := fmt.Sprintf("%s %s;%s %s;%s %s;%s %s;%s %s", dom.CSS_PARAM_DISPLAY, dom.CSS_PARAM_DISPLAY_FLEX,
		dom.CSS_PARAM_FLOW, flow, dom.CSS_PARAM_JUSTIFY, justify, dom.CSS_PARAM_ALIGN, align, dom.CSS_PARAM_GAP, gapStr)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, layout) }
}

func GridLayout[T string | int | int32 | int64 | float32 | float64](columns T, rows T, gap string) DomComponent {
	columnsStr := AnyStr(columns)
	rowsStr := AnyStr(rows)
	sanitizeHtml(&columnsStr)
	sanitizeHtml(&rowsStr)
	sanitizeHtml(&gap)
	layout := fmt.Sprintf("%s %s;%s %s%s%s;%s %s%s%s;%s %s", dom.CSS_PARAM_DISPLAY, dom.CSS_PARAM_DISPLAY_GRID,
		dom.CSS_PARAM_GRID_COLUMNS, dom.CSS_PARAM_GRID_REPEAT_OPENER, columnsStr, dom.CSS_PARAM_GRID_REPEAT_CLOSER,
		dom.CSS_PARAM_GRID_ROWS, dom.CSS_PARAM_GRID_REPEAT_OPENER, rowsStr, dom.CSS_PARAM_GRID_REPEAT_CLOSER, dom.CSS_PARAM_GAP, gap)
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_STYLE, layout) }
}

// DomComponentsParamsBinding

func OnClick(callbacks ...func(js.Value)) DomComponent {
	id := utils.GenerateShortId()
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_CLICK, nil, callbacks...))
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_ID, id) }
}

func OnChange(value *any, callbacks ...func(js.Value)) DomComponent {
	id := utils.GenerateShortId()
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_CHANGE, value, callbacks...))
	bindings[id] = append(bindings[id], generateBinding(dom.JS_EVENT_KEYUP, value, callbacks...))
	return func() string { return fmt.Sprintf("%s%s'%s'", dom.ELEMENT_PARAM, dom.HTML_PARAM_ID, id) }
}

// External utils

func AnyStr(v any) string {
	return fmt.Sprintf("%v", v)
}

func AnyInt(v any) int {
	x, _ := strconv.Atoi(fmt.Sprintf("%v", v))
	return x
}

func AnyFlt(v any) float64 {
	x, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	return x
}

func AnyBol(v any) bool {
	x, _ := strconv.ParseBool(fmt.Sprintf("%v", v))
	return x
}
