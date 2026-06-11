package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wii/grepom/service"
)

const minNameWidth = 4 // minimum column width for "NAME" header

type viewMode int

const (
	viewList viewMode = iota
	viewLogs
	viewDetail
)

type model struct {
	mgr       *service.Manager
	entries   []service.Entry
	cursor    int
	mode      viewMode
	logLines  []string
	logOffset int64
	message   string
	width     int
	height    int
	quitting  bool
}

type logLinesMsg struct {
	lines  []string
	offset int64
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
	m.logOffset = service.LogFileSize(entry.Record.LogPath)
	m.mode = viewLogs
	return nil
}

func (m model) scheduleLogFollow() tea.Cmd {
	entry := m.selected()
	if entry == nil || m.mode != viewLogs {
		return nil
	}
	path := entry.Record.LogPath
	offset := m.logOffset
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		lines, newOffset, err := service.ReadLinesFromOffset(path, offset)
		if err != nil {
			return logLinesMsg{offset: offset}
		}
		return logLinesMsg{lines: lines, offset: newOffset}
	})
}

func (m *model) appendLogLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	m.logLines = append(m.logLines, lines...)
	maxLines := m.height - 4
	if maxLines < 20 {
		maxLines = 20
	}
	if len(m.logLines) > maxLines {
		m.logLines = m.logLines[len(m.logLines)-maxLines:]
	}
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

func (m *model) restart() error {
	entry := m.selected()
	if entry == nil {
		return fmt.Errorf("no service selected")
	}
	rec, err := m.mgr.Restart(entry.Record.Name)
	if err != nil {
		return err
	}
	m.message = fmt.Sprintf("restarted %s (pid %d)", rec.Name, rec.PID)
	return m.refresh()
}

func (m model) servicePath() string {
	entry := m.selected()
	if entry == nil {
		return ""
	}
	return entry.Record.Cwd
}

// nameWidth calculates the column width for the NAME field so that all
// entries align correctly even when service names are longer than the header.
func (m model) nameWidth() int {
	w := minNameWidth
	for _, e := range m.entries {
		if len(e.Record.Name) > w {
			w = len(e.Record.Name)
		}
	}
	return w
}

func (m model) listView() string {
	var b strings.Builder
	b.WriteString("grepom svc tui  [j/k move  l logs  s stop  S kill-9  R restart  c clean  p path  r refresh  q quit]\n\n")
	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n\n")
	}
	if len(m.entries) == 0 {
		b.WriteString("No services found.\n")
		return b.String()
	}
	nw := m.nameWidth()
	b.WriteString(fmt.Sprintf("  %-*s %-8s %-8s %s\n", nw, "NAME", "STATUS", "PID", "PATH"))
	for i, e := range m.entries {
		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		pid := "-"
		if e.Record.PID > 0 {
			pid = fmt.Sprintf("%d", e.Record.PID)
		}
		b.WriteString(fmt.Sprintf("%s %-*s %-8s %-8s %s\n", marker, nw, e.Record.Name, e.Status, pid, e.Record.Cwd))
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
	b.WriteString("  [following  b back  q quit]\n\n")
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
