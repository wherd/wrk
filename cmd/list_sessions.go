package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Operation int64

const (
	Noop Operation = iota
	Activate
	Delete
)

type sessionModel struct {
	choices  []string
	cursor   int
	selected int
	action   Operation
}

func (m sessionModel) Init() tea.Cmd {
	return nil
}

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			m.action = Noop
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.action = Activate
			m.selected = m.cursor
			return m, tea.Quit

		case "-":
			m.action = Delete
			m.selected = m.cursor
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m sessionModel) View() string {
	s := "What session you want to activate/delete?\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = " ❯"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit, - to delete, <enter> to activate.\n"

	return s
}

var listSessionsCommand = &cobra.Command{
	Use:   "sl",
	Short: "List sessions",
	Long:  "List and manage project sessions.",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		out, err := exec.Command(path, "stash", "list").Output()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		if string(out) == "" {
			fmt.Println("No sessions found")
			os.Exit(0)
		}

		options := strings.Split(strings.Trim(string(out), "\n"), "\n")
		m := sessionModel{
			choices:  options,
			cursor:   0,
			selected: 0,
			action:   Noop,
		}

		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		if m, ok := result.(sessionModel); ok && m.action != Noop {

			switch m.action {

			case Activate:
				err = exec.Command(path, "stash", "pop", "--index", strconv.Itoa(m.selected)).Run()
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}

			case Delete:
				err = exec.Command(path, "stash", "drop", strconv.Itoa(m.selected)).Run()
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
			}
		}
	},
}
