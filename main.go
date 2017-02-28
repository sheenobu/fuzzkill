package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/tncardoso/gocurses"
)

func main() {

	var res string
	var ret int

	defer func() {
		os.Exit(ret)
	}()

	defer fmt.Fprintf(os.Stderr, "%s\n", res)

	if len(os.Args) < 2 {
		ret = -1
		return
	}

	arg := os.Args[1]

	gocurses.Initscr()
	defer gocurses.End()

	defer gocurses.Clear()

	gocurses.Cbreak()
	gocurses.Noecho()
	gocurses.Stdscr.Keypad(true)
	//wind := gocurses.NewWindow(0, 0, 1, 0)
	//wind.Box(0, 0)
	//wind.Refresh()
	var processes []struct {
		PID string
		Row int
	}

	var selected int
	var running = true

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	go func() {
		<-ch
		running = false
	}()

	var render = func() {

		cmd := exec.Command("ps", "-eo", "pid,fname,command", "-a", "-x")
		out, err := cmd.Output()
		if err != nil {
			res = fmt.Sprintf("Error listing processes: %s\n", err)
			ret = -1
			running = false
			return
		}

		processes = []struct {
			PID string
			Row int
		}{}

		gocurses.Clear()
		gocurses.Addstr("fuzzkill - searching '" + arg + "'\n")
		gocurses.Addstr("- 'q' to exit, 'k' to kill, 'tab' to select item\n")
		gocurses.Refresh()

		var l []string
		lines := strings.Split(string(out), "\n")
		var i int

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			tokens := strings.Split(line, " ")

			cname := tokens[1]
			if strings.Contains(cname, arg) {

				l = append(l, line)
				var prefix string
				if selected == i {
					prefix = "* "
					gocurses.Attron(gocurses.A_BOLD)
				}

				gocurses.Addstr(fmt.Sprintf("%s\t%s\t'%s'\t'%s'\n", prefix, tokens[0], tokens[1], strings.TrimSpace(strings.Join(tokens[2:], " ")[0:60])))
				processes = append(processes, struct {
					PID string
					Row int
				}{tokens[0], i})

				gocurses.Attroff(gocurses.A_BOLD)

				i++
			}
		}

		if len(processes) == 0 {
			ret = 0
			running = false
			return
		}
	}

	render()

	for running {
		ch := gocurses.Stdscr.Getch()
		switch ch {
		case 'q':
			running = false
		case '\t':
			selected++
			if selected >= len(processes) {
				selected = 0
			}
		case 'k':
			pid, err := strconv.Atoi(processes[selected].PID)
			if err != nil {
				res = err.Error()
				ret = -1
				return
			}

			proc, err := os.FindProcess(pid)
			if err != nil {
				res = err.Error()
				ret = -1
				return
			}

			err = proc.Signal(syscall.SIGTERM)
			if err != nil {
				res = err.Error()
				ret = -1
				return
			}
		}

		render()
	}

}
