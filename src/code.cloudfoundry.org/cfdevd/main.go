package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/cfdevd/cmd"
	"code.cloudfoundry.org/cfdevd/launchd"
)

const SockName = "ListenSocket"

func handleRequest(conn *net.UnixConn) {
	if err := doHandshake(conn); err != nil {
		fmt.Println("Handshake Error: ", err)
		return
	}
	command, err := cmd.UnmarshalCommand(conn)
	if err != nil {
		fmt.Println("Command:", err)
		return
	}
	command.Execute(conn)
}

func registerSignalHandler() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		os.Exit(0)
	}(sigc)
}

func install(programSrc string) {
	lctl := launchd.New()
	program := "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd"
	cfdevdSpec := launchd.DaemonSpec{
		Label:   "org.cloudfoundry.cfdevd",
		Program: program,
		ProgramArguments: []string{
			program,
		},
		RunAtLoad: false,
		Sockets: map[string]string{
			SockName: "/var/tmp/cfdevd.socket",
		},
		StdoutPath: "/var/tmp/cfdevd.stdout.log",
		StderrPath: "/var/tmp/cfdevd.stderr.log",
	}
	if err := lctl.AddDaemon(cfdevdSpec, programSrc); err != nil {
		fmt.Println("Failed to install cfdevd: ", err)
	}
}

func uninstall(prog string) {
	lctl := launchd.New()
	cfdevdSpec := launchd.DaemonSpec{
		Label:   "org.cloudfoundry.cfdevd",
		Program: "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd",
	}
	if err := lctl.RemoveDaemon(cfdevdSpec); err != nil {
		fmt.Println("Failed to uninstall cfdevd: ", err)
	}
}

func run() {
	registerSignalHandler()
	listeners, err := launchd.Listeners(SockName)
	if err != nil || len(listeners) != 1 {
		log.Fatal("Failed to obtain socket from launchd")
	}
	listener, ok := listeners[0].(*net.UnixListener)
	if !ok {
		log.Fatal("Failed to cast listener to unix listener")
	}
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			continue
		}
		defer conn.Close()
		go handleRequest(conn)
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			install(os.Args[0])
		case "uninstall":
			uninstall(os.Args[0])
		default:
			log.Fatal("unrecognized command ", os.Args[1])
		}
	} else {
		run()
	}
}
