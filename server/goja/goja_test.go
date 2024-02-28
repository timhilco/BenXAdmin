package goja

import (
	"benefitsDomain/domain/person"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

func Test_Goja1(t *testing.T) {
	vm := goja.New()
	v, err := vm.RunString("2 + 2")
	if err != nil {
		panic(err)
	}
	if num := v.Export().(int64); num != 4 {
		panic(num)
	}
}
func Test_Goja3(t *testing.T) {
	type Field struct {
	}
	type S struct {
		Field *Field
	}
	var s = S{
		Field: &Field{},
	}
	vm := goja.New()
	vm.Set("s", &s)
	res, err := vm.RunString(`
	var sym = Symbol(66);
	var field1 = s.Field;
	field1[sym] = true;
	var field2 = s.Field;
	field1 === field2; // true, because the equality operation compares the wrapped values, not the wrappers
	field1[sym] === true; // true
	field2[sym] === undefined; // also true
	return field1;
	`)
	if err != nil {
		fmt.Printf("%v", err)

	}
	fmt.Printf("%v", res)
}
func Test_Goja4(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	script := `
			console.log("Hello world - from Javascript inside Go! ")
		`
	fmt.Println("Compiling ... ")
	prog, err := goja.Compile("", script, true)
	if err != nil {
		fmt.Printf("Error compiling the script %v ", err)
		return
	}
	fmt.Println("Running ... \n ")
	_, err = vm.RunProgram(prog)
	if err != nil {
		panic(err)
	}

}

func Test_Goja2(t *testing.T) {
	script := getScript("./script1.js")
	vm := goja.New()
	v, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	if num := v.Export().(int64); num != 4 {
		panic(num)
	}
}
func getScript(fileName string) string {
	domain, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return string(domain)
}

func Test_Goja5(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	script := `
			function myFunction(param)
			{
				console.log("myFunction running ...")
				console.log("Param = ", param)
				return "Nice meeting you, Go"
			}
		`

	prog, err := goja.Compile("", script, true)
	if err != nil {
		fmt.Printf("Error compiling the script %v ", err)
		return
	}
	_, _ = vm.RunProgram(prog)

	var myJSFunc goja.Callable
	err = vm.ExportTo(vm.Get("myFunction"), &myJSFunc)
	if err != nil {
		fmt.Printf("Error exporting the function %v", err)
		return
	}

	res, err := myJSFunc(goja.Undefined(), vm.ToValue("message from go"))
	if err != nil {
		fmt.Printf("Error calling function %v", err)
		return
	}
	fmt.Printf("Returned value from JS function\n%s \n", res.ToString())

}

type Hooks struct {
	OnNewEmail         []*goja.Value
	BeforeSendingEmail []func(e *Email)
}

type Email struct {
	Subject  string
	Body     string
	Priority int
	To       []string
}

func (h *Hooks) Init() {
	h.OnNewEmail = make([]*goja.Value, 0)
	h.BeforeSendingEmail = make([]func(e *Email), 0)
}

func (h *Hooks) TriggerNewEmailEvent(email *Email, vm *goja.Runtime) {

	eobj := makeEmailJSObject(vm, email)

	for _, newEmail := range h.OnNewEmail {
		var newEmailCallBack func(*goja.Object)
		vm.ExportTo(*newEmail, &newEmailCallBack)
		newEmailCallBack(eobj)
	}
}

func makeEmailJSObject(vm *goja.Runtime, email *Email) *goja.Object {
	obj := vm.NewObject()
	obj.Set("subject", email.Subject)
	obj.Set("body", email.Body)
	obj.Set("to", email.To)
	obj.Set("reply", func(body string) {
		fmt.Printf("Replying:\n%s\n", body)
	})
	obj.Set("setPriority", func(p int) {
		fmt.Printf("Set email priority to %d\n", p)
	})

	obj.Set("moveTo", func(folder string) {
		fmt.Printf("Moving email to folder %s\n", folder)
	})

	return obj
}
func Test_Goja6(t *testing.T) {
	var hooks Hooks
	hooks.Init()

	vm := goja.New()

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	obj := vm.NewObject()

	obj.Set("RegisterHook", func(hook string, fn goja.Value) {
		switch hook {
		case "onEmailReceived":
			hooks.OnNewEmail = append(hooks.OnNewEmail, &fn)
			fmt.Println("Registered onEmailReceived Hook ")
		}

	})

	vm.Set("myemail", obj)

	script := `
		console.log("JS code started ")
		
		myemail.RegisterHook("onEmailReceived", iGotEmail)

		function iGotEmail(newEmail)
		{
			console.log("newEmail, subject %s ", newEmail.subject,  newEmail.to)
			
			if(newEmail.subject.startsWith("URGENT:"))
			{
				newEmail.setPriority(5)
				
				newEmail.reply("Hello,\n Received your email. We will respond on priority basis. \n\nThanks\n")
				return
			}
			else if(newEmail.to.includes("sales@website"))
			{
				newEmail.moveTo("Sales")
				return
			}
		}
	`
	prg, err := goja.Compile("", script, true)
	if err != nil {
		fmt.Printf("Error compiling the script %v ", err)
		return
	}

	_, _ = vm.RunProgram(prg)

	email1 := &Email{
		Subject: "URGENT: Systems down!",
		Body:    "5 of your systems are down at the moment",
		To:      []string{"some@one.cc"},
	}

	fmt.Println("Triggering the urgent email event ")
	hooks.TriggerNewEmailEvent(email1, vm)

	email2 := &Email{
		Subject: "New order received!",
		Body:    "You got new order for 1k blue widgets!",
		To:      []string{"sales@website"},
	}

	fmt.Println("Triggering the sales email event ")
	hooks.TriggerNewEmailEvent(email2, vm)
}

func Test_Goja7(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	person := person.Person{
		FirstName: "John",
		LastName:  "Sample",
	}
	vm.Set("person", person)
	s := getScript("./script1.js")
	_, err := vm.RunString(s)
	if err != nil {
		panic(err)
	}
	handle, ok := goja.AssertFunction(vm.Get("handle"))
	if !ok {
		panic("Not a function")
	}

	res, err := handle(goja.Undefined())
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

}
