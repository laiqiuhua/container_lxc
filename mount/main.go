package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func createTxtFile() {
	f, err := os.Create("/tmp/leopard.txt")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.WriteString("leopard")
	if err != nil {
		panic(err)
	}
}

func execContainerShell() {
	log.Printf("Ready to exec container shell ...\n")

	if err := syscall.Sethostname([]byte("leopard")); err != nil {
		panic(err)
	}

	log.Printf("Chaning to /tmp directory ...\n")

	if err := os.Chdir("/tmp"); err != nil {
		panic(err)
	}

	log.Printf("Mounting / as private ...\n")

	mf := uintptr(syscall.MS_PRIVATE | syscall.MS_REC)
	if err := syscall.Mount("", "/", "", mf, ""); err != nil {
		panic(err)
	}

	log.Printf("Binding rootfs/ to rootfs/ ...\n")

	mf = uintptr(syscall.MS_BIND | syscall.MS_REC)
	if err := syscall.Mount("rootfs/", "rootfs/", "", mf, ""); err != nil {
		panic(err)
	}

	log.Printf("Pivot new root at rootfs/ ...\n")

	if err := syscall.PivotRoot("rootfs/", "rootfs/.old_root"); err != nil {
		panic(err)
	}

	log.Printf("Changing to / directory ...\n")

	if err := os.Chdir("/"); err != nil {
		panic(err)
	}

	log.Printf("Mounting /tmp as tmpfs ...\n")

	mf = uintptr(syscall.MS_NODEV)
	if err := syscall.Mount("tmpfs", "/tmp", "tmpfs", mf, ""); err != nil {
		panic(err)
	}

	log.Printf("Mounting /proc filesystem ...\n")

	mf = uintptr(syscall.MS_NODEV)
	if err := syscall.Mount("proc", "/proc", "proc", mf, ""); err != nil {
		panic(err)
	}

	createTxtFile()

	log.Printf("Mounting /.old_root as private ...\n")

	mf = uintptr(syscall.MS_PRIVATE | syscall.MS_REC)
	if err := syscall.Mount("", "/.old_root", "", mf, ""); err != nil {
		panic(err)
	}

	log.Printf("Unmount parent rootfs from /.old_root ...\n")

	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
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
