package cmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
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
	Long:  "List (un)stage and commit files.",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		m := listFilesModel{
			choices: []string{},
			cursor:  0,
			commit:  false,
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

		result, err := p.Run()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		if m, ok := result.(listFilesModel); ok && m.commit {
			if len(m.staged) == 0 {
				fmt.Println("Nothing to commit")
				os.Exit(1)
			}

			str := ""
			for file, _ := range m.staged {
				str += file + "\n"
			}
			clipboard.Write(clipboard.FmtText, []byte(str))

			p := tea.NewProgram(createInputModel())
			result, err := p.Run()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}

			if m, ok := result.(inputModel); ok {
				str := m.textInput.Value()
				if str == "" {
					fmt.Println("Commit message is empty")
					os.Exit(1)
				}

				if err := run("commit", "-m", str); err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}

				if err := run("push"); err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
			}
		}
	},
}

type listFilesModel struct {
	choices []string
	cursor  int
	commit  bool
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
			m.commit = true
			return m, tea.Quit
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

	s += "\nPress q to quit, c to commit changes.\n"

	return s
}

type errMsg error

type inputModel struct {
	textInput textinput.Model
	err       error
}

func createInputModel() inputModel {
	ti := textinput.New()
	ti.Placeholder = "initial commit"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return inputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		"Commit message:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
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
