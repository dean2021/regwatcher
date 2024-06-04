package main

import "github.com/dean2021/regwatcher"

func main() {
	path := `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	watcher, err := regwatcher.NewWatcher(regwatcher.HKeyLocalMachine, path, 1000)
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	for {
		changed, err := watcher.Watch()
		if err != nil {
			panic(err)
		}
		if changed {
			println("registry changed")
		}
	}
}
