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

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
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
		os.Exit(0)
	}

	log.Printf("Ready to run command ...\n")

	cmd := exec.Command(os.Args[0], []string{clone}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
