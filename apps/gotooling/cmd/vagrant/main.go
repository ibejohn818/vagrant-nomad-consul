package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func main() {

	cwd, _ := filepath.Abs("../../")
	cmd := exec.Command("vagrant", "status", "--machine-readable")
	cmd.Dir = cwd

  fmt.Println("Dir: ", cmd.Dir)

	cmd.Run()
	cmd.Wait()

	out, _ := cmd.CombinedOutput()
	fmt.Println("Out: ", out)

}
