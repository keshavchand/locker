package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Config struct {
	LogIPCommands bool
}

var config Config

func main() {
	config.LogIPCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Locker: ")
	childFlag := flag.Bool("child", false, "run as child")
	flag.Parse()

	if childFlag == nil || *childFlag == false {
		parent()
	} else if *childFlag == true {
		child()
	}
}

func child() {
	pid := os.Getpid()
	cwd, err := os.Getwd()
	must(err)
	log.Printf("Running as child process with pid: %d @ %s", pid, cwd)
	must(syscall.Chroot("rootfs"))
	must(os.Chdir("/"))
	must(os.MkdirAll("/proc", 0755))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// dirTree("/busybox sh", 0)

	// lookup google.com on [::1]:53: read udp [::1]:53225->[::1]:53: read: connection refused
	// runCmd(config.LogIPCommands, "ip", "link", "set", "lo", "up")
	// runCmd(config.LogIPCommands, "ip", "addr", "add", "192.168.1.1/24", "dev", "lo")

	// create bridge
	// runCmd(config.LogIPCommands, "ip", "link", "add", "name", "br0", "type", "bridge")
	// runCmd(config.LogIPCommands, "ip", "link", "set", "dev", "br0", "up")
	// runCmd(config.LogIPCommands, "ip", "addr", "add", "192.168.1.1/24", "dev", "br0")

	// NOTE: the binary must be a static compiled binary
	// Go binaries should be compiled with
	// -tags netgo -ldflags '-extldflags "-static"'
	location := "./main"
	cmd := exec.CommandContext(context.TODO(), location, "sh")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	log.Println("Starting Process with pid: ", cmd.Process.Pid)
	must(cmd.Wait())
	must(syscall.Unmount("/proc", 0))
}

func parent() {
	location := "/proc/self/exe"
	cmd := exec.CommandContext(context.TODO(), location, "-child")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER, // | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
	}
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func runCmd(should_log bool, cmd string, args ...string) {
	if should_log {
		log.Println("Running command:", cmd, args)
	}
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	must(c.Run())
}

func dirTree(start string, level int) {
	dirs, err := os.ReadDir(start)
	if err != nil {
		return
	}

	for _, dir := range dirs {
		for i := 0; i < level; i++ {
			fmt.Print("  ")
		}
		fmt.Println(dir.Name())
		dirTree(dir.Name(), level+1)
	}
}
