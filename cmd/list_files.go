package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
	"os"
	"os/exec"
	"slices"
	"strings"
)

var listFilesCommand = &cobra.Command{
	Use:   "fl",
	Short: "List files",
	Long:  "List modified files and (un)stage.",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		m := listFilesModel{
			choices: []string{},
			cursor:  0,
			staged:  map[string]bool{},
		}

		if out, err := exec.Command(path, "status", "-u", "--porcelain").Output(); err == nil {
			str := strings.Trim(string(out), "\n")
			if str == "" {
				fmt.Println("No files changed")
				os.Exit(0)
			}

			m.choices = strings.Split(str, "\n")
			for i, file := range m.choices {
				m.choices[i] = string(file[3:])
			}
		} else {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		if out, err := exec.Command(path, "diff", "--name-only", "--cached").Output(); err == nil {
			str := strings.Trim(string(out), "\n")
			if str != "" {
				staged := strings.Split(str, "\n")
				for _, file := range staged {
					if idx := slices.Index(m.choices, file); idx != -1 {
						m.staged[file] = true
					}
				}
			}
		} else {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

type listFilesModel struct {
	choices []string
	cursor  int
	staged  map[string]bool
}

func (m listFilesModel) Init() tea.Cmd {
	return nil
}

func (m listFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
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
			file := m.choices[m.cursor]
			if _, ok := m.staged[file]; ok {
				if err := run("restore", "--staged", file); err == nil {
					delete(m.staged, file)
				}
			} else {
				if err := run("add", file); err == nil {
					m.staged[file] = true
				}
			}

		case "d":
			file := m.choices[m.cursor]
			if _, ok := m.staged[file]; ok {
				if err := run("restore", "--staged", file); err == nil {
					delete(m.staged, file)
				}
			}

			if err := run("restore", file); err == nil {
				m.choices = append(m.choices[:m.cursor], m.choices[m.cursor+1:]...)
			}

		case "c":
			{
				str := ""
				for file, _ := range m.staged {
					str += file + "\n"
				}
				clipboard.Write(clipboard.FmtText, []byte(str))
			}
		}
	}

	return m, nil
}

func (m listFilesModel) View() string {
	s := "What files you want to (un)stage?\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = " ❯"
		}

		staged := " ○"
		if _, ok := m.staged[choice]; ok {
			staged = " ●"
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, staged, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}

func run(args ...string) error {
	path, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	if err := exec.Command(path, args...).Run(); err != nil {
		return err
	}

	return nil
}
