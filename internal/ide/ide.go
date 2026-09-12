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

const (
	prefNone   = ""
	prefVSCode = "vscode"
	prefCursor = "cursor"
	prefClaude = "claude"
)

// DoubleClickIDELabels are the labels shown in the default-editor selector.
var DoubleClickIDELabels = []string{"None", "VS Code", "Cursor", "Claude Code"}

// KindFromPref maps a stored preference value to an IDE kind.
// Returns false when no editor is configured (None).
func KindFromPref(value string) (Kind, bool) {
	switch value {
	case prefVSCode:
		return VSCode, true
	case prefCursor:
		return Cursor, true
	case prefClaude:
		return Claude, true
	default:
		return VSCode, false
	}
}

// PrefFromKind returns the stored preference value for an IDE kind.
func PrefFromKind(k Kind) string {
	switch k {
	case VSCode:
		return prefVSCode
	case Cursor:
		return prefCursor
	case Claude:
		return prefClaude
	default:
		return prefNone
	}
}

// LabelFromPref returns the UI label for a stored preference value.
func LabelFromPref(value string) string {
	switch value {
	case prefVSCode:
		return "VS Code"
	case prefCursor:
		return "Cursor"
	case prefClaude:
		return "Claude Code"
	default:
		return "None"
	}
}

// PrefFromLabel returns the stored preference value for a UI label.
func PrefFromLabel(label string) string {
	switch label {
	case "VS Code":
		return prefVSCode
	case "Cursor":
		return prefCursor
	case "Claude Code":
		return prefClaude
	default:
		return prefNone
	}
}
