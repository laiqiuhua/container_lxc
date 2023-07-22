package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func execContainerShell() {
	log.Printf("Ready to exec container shell ...\n")

	if err := syscall.Sethostname([]byte("leopard")); err != nil {
		panic(err)
	}

	const sh = "/bin/sh"

	env := os.Environ()
	env = append(env, "PS1=-> ")

	if err := syscall.Exec(sh, []string{""}, env); err != nil {
		panic(err)
	}
}

func main() {
	log.Printf("Starting process %s with args: %v\n", os.Args[0], os.Args)

	const clone = "CLONE"

	if len(os.Args) > 1 && os.Args[1] == clone {
		execContainerShell()
	}

	log.Printf("Ready to run command ...\n")

	cmd := exec.Command(os.Args[0], []string{clone}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
