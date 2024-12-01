package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/manifoldco/promptui"
)

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func chooseJobMenu(menu *jobMenu) {

}

func homeMenu(menu *jobMenu) {
	type option struct {
		name  string
		state jobMenuState
	}

	ops := make([]*option, 0)

	ops = append(ops, &option{
		name:  "Choose Jobs",
		state: jobMenuStateChooseJob,
	})

	if len(menu.jobs) > 0 {
		ops = append(ops, &option{
			name:  "Run Jobs",
			state: jobMenuStateRunJob,
		})
	}
	ops = append(ops, &option{
		name:  "Quit",
		state: jobMenuStateExit,
	})

	items := make([]string, 0)
	for _, v := range ops {
		items = append(items, v.name)
	}

	prompt := promptui.Select{
		Label: "Choose Option",
		Items: items,
	}

	_, res, err := prompt.Run()

	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}

	for _, v := range ops {
		if res == v.name {
			menu.state = v.state
			return
		}
	}

	menu.state = jobMenuStateExit

	spew.Dump("Res: ", res)
	spew.Dump("Err: ", err)

}

type HclVarType string

const (
	HclVarType_String HclVarType = "string"
	HclVarType_Number HclVarType = "number"
)

type HclVarInput struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Type  HclVarType  `json:"type"`
}

type JobInput struct {
	HclPath string         `json:"hcl_path"`
	HclVars []*HclVarInput `json:"hl_vars"`
}

type jobMenuState string

const (
	jobMenuStateHome      jobMenuState = "home"
	jobMenuStateChooseJob jobMenuState = "choose-job"
	jobMenuStateRunJob    jobMenuState = "run-jobs"
	jobMenuStateExit      jobMenuState = "exit"
)

type jobMenu struct {
	state      jobMenuState
	jobs       []*JobInput
	jobHclPath string
}

func clearTerm() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
