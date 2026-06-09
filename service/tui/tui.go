package tui

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wii/grepom/service"
)

// Run starts the interactive service management UI.
func Run(ctx context.Context, mgr *service.Manager) error {
	if ctx == nil {
		ctx = context.Background()
	}
	m := newModel(mgr)
	if err := m.refresh(); err != nil {
		return err
	}
	p := tea.NewProgram(modelWithContext{model: m, ctx: ctx}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

type modelWithContext struct {
	model
	ctx context.Context
}

func (m modelWithContext) Init() tea.Cmd {
	return nil
}

func (m modelWithContext) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case viewLogs, viewDetail:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "b", "esc":
				m.mode = viewList
				return m, nil
			}
		default:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "j", "down":
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "r":
				_ = m.refresh()
			case "l":
				if err := m.loadLogs(); err != nil {
					m.message = err.Error()
				}
			case "d":
				m.mode = viewDetail
			case "p":
				if path := m.servicePath(); path != "" {
					m.message = path
				}
			case "s":
				if err := m.kill(false); err != nil {
					m.message = err.Error()
				}
			case "S":
				if err := m.kill(true); err != nil {
					m.message = err.Error()
				}
			case "c":
				if _, err := m.clean(); err != nil {
					m.message = err.Error()
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m modelWithContext) View() string {
	if m.quitting {
		return ""
	}
	return m.model.View()
}

// PrintServicePath writes the selected service path to stdout for shell integration tests.
func PrintServicePath(mgr *service.Manager, name string) error {
	path, err := mgr.Dir(name)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

// EnsureTTY reports whether stdin is a terminal.
func EnsureTTY() error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("svc tui requires a TTY")
	}
	return nil
}
