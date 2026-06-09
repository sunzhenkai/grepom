package tui

import (
	"fmt"
	"strings"

	"github.com/wii/grepom/service"
)

type viewMode int

const (
	viewList viewMode = iota
	viewLogs
	viewDetail
)

type model struct {
	mgr      *service.Manager
	entries  []service.Entry
	cursor   int
	mode     viewMode
	logLines []string
	message  string
	width    int
	height   int
	quitting bool
}

func newModel(mgr *service.Manager) model {
	return model{mgr: mgr, mode: viewList}
}

func (m model) selected() *service.Entry {
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return nil
	}
	return &m.entries[m.cursor]
}

func (m *model) refresh() error {
	entries, err := m.mgr.List()
	if err != nil {
		return err
	}
	m.entries = entries
	if m.cursor >= len(m.entries) {
		if len(m.entries) == 0 {
			m.cursor = 0
		} else {
			m.cursor = len(m.entries) - 1
		}
	}
	return nil
}

func (m *model) loadLogs() error {
	entry := m.selected()
	if entry == nil {
		return fmt.Errorf("no service selected")
	}
	lines, err := service.ReadTailLines(entry.Record.LogPath, 30)
	if err != nil {
		return err
	}
	m.logLines = lines
	m.mode = viewLogs
	return nil
}

func (m *model) kill(force bool) error {
	entry := m.selected()
	if entry == nil {
		return fmt.Errorf("no service selected")
	}
	if err := m.mgr.Kill(entry.Record.Name, force); err != nil {
		return err
	}
	m.message = fmt.Sprintf("stopped %s", entry.Record.Name)
	return m.refresh()
}

func (m *model) clean() (int, error) {
	removed, err := m.mgr.Clean(service.CleanOptions{})
	if err != nil {
		return 0, err
	}
	m.message = fmt.Sprintf("cleaned %d service(s)", removed)
	if err := m.refresh(); err != nil {
		return removed, err
	}
	return removed, nil
}

func (m model) servicePath() string {
	entry := m.selected()
	if entry == nil {
		return ""
	}
	return entry.Record.Cwd
}

func (m model) listView() string {
	var b strings.Builder
	b.WriteString("grepom svc tui  [j/k move  l logs  s stop  S kill-9  c clean  p path  r refresh  q quit]\n\n")
	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n\n")
	}
	if len(m.entries) == 0 {
		b.WriteString("No services found.\n")
		return b.String()
	}
	b.WriteString(fmt.Sprintf("%-16s %-8s %-8s %s\n", "NAME", "STATUS", "PID", "PATH"))
	for i, e := range m.entries {
		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		pid := "-"
		if e.Record.PID > 0 {
			pid = fmt.Sprintf("%d", e.Record.PID)
		}
		b.WriteString(fmt.Sprintf("%s%-15s %-8s %-8s %s\n", marker, e.Record.Name, e.Status, pid, e.Record.Cwd))
	}
	return b.String()
}

func (m model) logsView() string {
	var b strings.Builder
	entry := m.selected()
	title := "logs"
	if entry != nil {
		title = fmt.Sprintf("logs: %s", entry.Record.Name)
	}
	b.WriteString(title)
	b.WriteString("  [b back  q quit]\n\n")
	for _, line := range m.logLines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) detailView() string {
	entry := m.selected()
	if entry == nil {
		return "no service selected"
	}
	return fmt.Sprintf(
		"service: %s\nstatus: %s\npid: %d\npath: %s\ncommand: %s\nlog: %s\n\n[b back]",
		entry.Record.Name,
		entry.Status,
		entry.Record.PID,
		entry.Record.Cwd,
		entry.Record.Command,
		entry.Record.LogPath,
	)
}

func (m model) View() string {
	switch m.mode {
	case viewLogs:
		return m.logsView()
	case viewDetail:
		return m.detailView()
	default:
		return m.listView()
	}
}
