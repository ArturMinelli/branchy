package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/git"
	"branchy/internal/project"
)

type remoteUpdateMsg struct {
	projectID string
	path      string
	err       error
}

func remoteUpdateCmd(p *project.Project) tea.Cmd {
	if p == nil {
		return nil
	}
	id, path := p.ID, p.Path
	return func() tea.Msg {
		return remoteUpdateMsg{
			projectID: id,
			path:      path,
			err:       git.FetchDefaultRemote(path),
		}
	}
}
