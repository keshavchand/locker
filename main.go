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
	TargetDir     string
}

var config Config

func main() {
	config.LogIPCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Locker: ")
	childFlag := flag.Bool("child", false, "run as child")
	chrootDir := flag.String("targetdir", "rootfs", "targetdir directory")
	flag.Parse()

	config.TargetDir = *chrootDir

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

	must(syscall.Chroot(config.TargetDir))
	must(os.Chdir("/"))

	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		log.Println("/proc directory does not exist, creating it")
		must(os.Mkdir("/proc", 0755))
		defer os.RemoveAll("/proc")
	}

	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// NOTE: the binary must be a static compiled binary
	// Go binaries should be compiled with
	// -tags netgo -ldflags '-extldflags "-static"'
	// XXX: Extrypoint should be provided by user
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
	args := os.Args[1:]
	args = append(args, "--child")
	cmd := exec.CommandContext(context.TODO(), location, args...)
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
