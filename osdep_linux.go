package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func runGameServer() {
	exeCmd = exec.Command(exePath, "-WorldID", config["lifds"]["world-id"])
	exeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr
	err := exeCmd.Start()

	if err != nil {
		log.Printf("Error in server start: %s", err)
		gameSrvStatusChan <- "DOWN (FAILED TO START)"
	}

	gameSrvStatusChan <- "UP"
	doneChan <- exeCmd.Wait()
}
