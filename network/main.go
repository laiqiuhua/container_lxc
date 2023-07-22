package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"
)

const (
	Bridge   = "br0"
	BridgeIp = "172.20.1.2/24"
	Lo       = "lo"
	Peer0    = "veth0"
	Peer0Ip  = "172.20.1.3/24"
	Peer1    = "veth1"
	Peer1Ip  = "172.20.1.4/24"
)

func createTxtFile() {
	f, err := os.Create("/tmp/leopard.txt")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString("leopard")
	if err != nil {
		panic(err)
	}

	_ = f.Close()
}

func checkBridge() (*netlink.Bridge, error) {
	la := netlink.NewLinkAttrs()
	la.Name = Bridge

	br := &netlink.Bridge{LinkAttrs: la}

	if _, err := net.InterfaceByName(Bridge); err != nil {
		return br, err
	}

	return br, nil
}

func setupBridge() error {
	br, err := checkBridge()
	if err != nil {
		log.Printf("Bridge %s does not exists ...\n", Bridge)
		log.Printf("Creating the Bridge %s ...\n", Bridge)

		if err = netlink.LinkAdd(br); err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		log.Printf("Bridge %s already exists ...\n", Bridge)
	}

	addr, err := netlink.ParseAddr(BridgeIp)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Attaching address %s to the Bridge %s ...\n", BridgeIp, Bridge)

	if err = netlink.AddrAdd(br, addr); err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Activating the Bridge %s ...\n", Bridge)

	if err = netlink.LinkSetUp(br); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func deleteBridge() error {
	br, err := checkBridge()
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Deactivating the Bridge %s ...\n", Bridge)

	if err := netlink.LinkSetDown(br); err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Deleting the Bridge %s ...\n", Bridge)

	if err := netlink.LinkDel(br); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func setupVethPeers() error {
	br, err := checkBridge()
	if err != nil {
		fmt.Println(err)
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = Peer0
	la.MasterIndex = br.Attrs().Index

	log.Printf("Creating the pairs %s and %s ...\n", Peer0, Peer1)

	// ip link add veth0 type veth peer name veth1
	veth := &netlink.Veth{LinkAttrs: la, PeerName: Peer1}
	if err := netlink.LinkAdd(veth); err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Link %s as master of %s ...\n", Bridge, Peer0)

	// ip link set veth0 master br0
	if err = netlink.LinkSetMaster(veth, br); err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Activating the pairs %s & %s ...\n", Peer0, Peer1)

	if err = netlink.LinkSetUp(veth); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func namespaceVethPeer(pid int) error {
	log.Printf("Getting the link for pair %s ...\n", Peer1)

	veth1, err := netlink.LinkByName(Peer1)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Namespacing the pair %s with pid %d ...\n", Peer1, pid)

	// ip link set veth1 netns $UPID
	if err := netlink.LinkSetNsPid(veth1, pid); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func activateLo() error {
	log.Printf("Getting the link for pair %s ...\n", Lo)

	loIf, err := netlink.LinkByName(Lo)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Activating %s ...\n", Lo)

	// ip link set dev lo up
	if err = netlink.LinkSetUp(loIf); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func activateVethPair(name, ip string) error {
	log.Printf("Getting the link for pair %s ...\n", name)

	veth, err := netlink.LinkByName(name)
	if err != nil {
		fmt.Println(err)
		return err
	}

	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Attaching address %s to the pair %s ...\n", ip, name)

	// ip addr add ip dev vethX
	if err = netlink.AddrAdd(veth, addr); err != nil {
		fmt.Println(err)
		return err
	}

	log.Printf("Activating the pair %s ...\n", name)

	// ip link set dev vethX up
	if err = netlink.LinkSetUp(veth); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
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

	if err := activateLo(); err != nil {
		panic(err)
	}

	if err := activateVethPair(Peer1, Peer1Ip); err != nil {
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
		// Clone
		execContainerShell()
	} else {
		// Parent
		if err := setupBridge(); err != nil {
			panic(err)
		}

		if err := setupVethPeers(); err != nil {
			panic(err)
		}

		if err := activateVethPair(Peer0, Peer0Ip); err != nil {
			panic(err)
		}
	}

	log.Printf("Ready to run command ...\n")

	cmd := exec.Command(os.Args[0], []string{clone}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	if err := namespaceVethPeer(cmd.Process.Pid); err != nil {
		panic(err)
	}

	_ = cmd.Wait()

	_ = deleteBridge()
}
