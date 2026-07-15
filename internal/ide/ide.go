package ide

// Kind identifies a supported IDE.
type Kind int

const (
	VSCode Kind = iota
	Cursor
	Claude
)

// Availability reports which IDEs appear installed on this machine.
type Availability struct {
	VSCode           bool
	Cursor           bool
	Claude           bool // Claude Code runtime (CLI or handler) — can open a worktree
	ClaudeDesktopApp bool // Claude.app chat/desktop client (Code tab needs Pro/Max)
}

// ClaudeHint returns the hover text for the Claude Code toolbar button.
func ClaudeHint(a Availability) string {
	if a.Claude {
		return "Open in Claude Code"
	}
	if a.ClaudeDesktopApp {
		return "Claude Code requires Pro/Max on desktop, or install the CLI: https://claude.ai/install"
	}
	return "Claude Code not installed"
}

// Name returns a human-readable IDE name.
func (k Kind) Name() string {
	switch k {
	case VSCode:
		return "VS Code"
	case Cursor:
		return "Cursor"
	case Claude:
		return "Claude Code"
	default:
		return "IDE"
	}
}
