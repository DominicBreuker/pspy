package dev

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/fsnotify/fsnotify"
)

type Process struct {
	pid   int
	ppid  int
	state rune
	pgrp  int
	sid   int

	binary string
}

func Monitor() {
	// procList := make(map[int]string)

	watch()

	// for {
	// 	refresh(procList)
	// }
}

func watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Can't create file system watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func refresh(procList map[int]string) error {
	proc, err := ioutil.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("opening proc dir: %v", err)
	}

	pids := make([]int, 0)

	for _, f := range proc {
		if f.IsDir() {
			name := f.Name()
			pid, err := strconv.Atoi(name)
			if err != nil {
				continue // not a pid
			}
			pids = append(pids, pid)
		}
	}

	for _, pid := range pids {
		_, ok := procList[pid]
		if !ok {
			cmd, err := getCmd(pid)
			if err != nil {
				cmd = "UNKNOWN" // process probably terminated
			}
			log.Printf("New process: %5d: %s\n", pid, cmd)
			procList[pid] = cmd
		}
	}
	return nil
}

func getCmd(pid int) (string, error) {
	cmdPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmd, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(cmd); i++ {
		if cmd[i] == 0 {
			cmd[i] = 32
		}
	}
	return string(cmd), nil
}
