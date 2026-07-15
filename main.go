package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jigarthacker24/git-worktree-manager/internal/gitops"
	"github.com/jigarthacker24/git-worktree-manager/internal/ide"
	"github.com/jigarthacker24/git-worktree-manager/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type appState struct {
	app        fyne.App
	prefs      fyne.Preferences
	window     fyne.Window
	repoPath   string
	worktrees  []gitops.Worktree
	branches   []string
	list       *widget.List
	colHeader  fyne.CanvasObject
	rowMetrics ui.RowMetrics
	status     *widget.Label
	selectedID widget.ListItemID
	ideAvail   ide.Availability
}

const (
	recentPathsKey      = "recent_repo_paths"
	pinnedWorktreesKey  = "pinned_worktrees"
	maxRecentPaths      = 5
	maxPinnedWorktrees  = 3
)

func main() {
	a := app.NewWithID("io.github.jigarthacker24.gitworktreemanager")
	w := a.NewWindow("Git Worktree Manager")
	w.SetMaster()

	state := &appState{
		app:        a,
		prefs:      a.Preferences(),
		window:     w,
		selectedID: -1,
	}
	w.SetContent(state.welcomeView())
	w.Show()
	fyne.Do(func() {
		ui.MaximizeWindow(w)
	})
	a.Run()
}

func (s *appState) welcomeView() fyne.CanvasObject {
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("/path/to/your/repo")

	recentPaths := s.loadRecentPaths()

	browseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			pathEntry.SetText(uri.Path())
		}, s.window)
	})

	openBtn := widget.NewButton("Open", func() {
		s.openRepo(strings.TrimSpace(pathEntry.Text))
	})
	openBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabel("Repository path"),
		container.NewBorder(nil, nil, nil, browseBtn, pathEntry),
		openBtn,
	)

	if len(recentPaths) > 0 {
		recentRows := container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabelWithStyle("Repository Dir", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle("Repository Path", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			),
		)
		for _, path := range recentPaths {
			recentRows.Add(s.recentRepoRow(path))
		}
		content.Add(widget.NewSeparator())
		content.Add(widget.NewLabel("Recent"))
		content.Add(recentRows)
	}

	return container.NewBorder(nil, nil, nil, nil, content)
}

func (s *appState) recentRepoRow(path string) fyne.CanvasObject {
	p := path
	return ui.NewTappableRow(filepath.Base(p), p, func() { s.openRepo(p) })
}

func (s *appState) loadRecentPaths() []string {
	return s.prefs.StringList(recentPathsKey)
}

func (s *appState) rememberPath(path string) {
	paths := s.prefs.StringList(recentPathsKey)
	updated := []string{path}
	for _, p := range paths {
		if p != path {
			updated = append(updated, p)
		}
	}
	if len(updated) > maxRecentPaths {
		updated = updated[:maxRecentPaths]
	}
	s.prefs.SetStringList(recentPathsKey, updated)
}

func (s *appState) openRepo(path string) {
	if path == "" {
		dialog.ShowError(fmt.Errorf("enter a repository path"), s.window)
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	if !gitops.IsRepo(abs) {
		dialog.ShowError(fmt.Errorf("not a git repository: %s", abs), s.window)
		return
	}
	s.repoPath = s.normalizedPath(abs)
	s.rememberPath(s.repoPath)
	s.window.SetContent(s.mainView())
	s.refresh()
}

func (s *appState) mainView() fyne.CanvasObject {
	s.status = widget.NewLabel("")
	s.ideAvail = ide.Detect()
	s.rowMetrics = ui.RowMetrics{DirWidth: ui.ComputeDirWidth(s.worktrees)}
	s.list = widget.NewList(
		func() int { return len(s.worktrees) },
		func() fyne.CanvasObject {
			copyIcon := func() *widget.Button {
				btn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)
				btn.Importance = widget.LowImportance
				return btn
			}
			pinBtn := widget.NewButtonWithIcon("", ui.PinIcon(), nil)
			pinBtn.Importance = widget.LowImportance

			branchLbl := widget.NewLabel("branch")
			pathLbl := widget.NewLabel("path")

			openBox := container.NewHBox(
				ui.IconButton(ui.VSCodeIcon(), "Open in VS Code", nil, s.setStatus),
				ui.IconButton(ui.CursorIcon(), "Open in Cursor", nil, s.setStatus),
				ui.IconButton(ui.ClaudeIcon(), "Open in Claude Code", nil, s.setStatus),
			)

			cols := ui.NewWorktreeCenter(&s.rowMetrics.DirWidth,
				widget.NewLabel("dir"),
				ui.NewTextWithCopy(branchLbl, copyIcon()),
				ui.NewTextWithCopy(pathLbl, copyIcon()),
			)
			return container.NewBorder(nil, nil, pinBtn, openBox, cols)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id == 0 {
				s.fitColumnsToRow(obj)
			}
			wt := s.worktrees[id]
			border := obj.(*fyne.Container)
			cols := border.Objects[0].(*fyne.Container)
			pinBtn := border.Objects[1].(*widget.Button)

			dirLbl := cols.Objects[0].(*widget.Label)
			branchCell := cols.Objects[1].(*fyne.Container)
			pathCell := cols.Objects[2].(*fyne.Container)
			openBox := border.Objects[2].(*fyne.Container)

			branchLbl := ui.LabelFromTextCell(branchCell)
			copyBranchBtn := ui.ButtonFromTextCell(branchCell)
			pathLbl := ui.LabelFromTextCell(pathCell)
			copyPathBtn := ui.ButtonFromTextCell(pathCell)
			ideBtns := ui.ButtonsFromHintHBox(openBox)

			dirLbl.SetText(wt.DirName)
			branchLbl.SetText(ui.WorktreeBranchLabel(wt))
			pathLbl.SetText(wt.Path)
			branchLbl.Truncation = fyne.TextTruncateOff
			if s.rowMetrics.TruncatePath {
				pathLbl.Truncation = fyne.TextTruncateEllipsis
			} else {
				pathLbl.Truncation = fyne.TextTruncateOff
			}

			if s.isPinned(wt.Path) {
				pinBtn.SetIcon(ui.PinFilledIcon())
			} else {
				pinBtn.SetIcon(ui.PinIcon())
			}
			pinBtn.OnTapped = func() {
				s.togglePin(wt.Path)
			}

			copyBranchBtn.OnTapped = func() {
				s.copyToClipboard(wt.Branch, "branch name")
			}
			copyPathBtn.OnTapped = func() {
				s.copyToClipboard(wt.Path, "worktree path")
			}

			s.bindIDEButton(ideBtns, 0, ide.VSCode, wt.Path, s.ideAvail.VSCode)
			s.bindIDEButton(ideBtns, 1, ide.Cursor, wt.Path, s.ideAvail.Cursor)
			s.bindIDEButton(ideBtns, 2, ide.Claude, wt.Path, s.ideAvail.Claude)
			if len(openBox.Objects) > 2 {
				ui.SetHint(openBox.Objects[2], ide.ClaudeHint(s.ideAvail))
			}
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		s.selectedID = id
	}

	addBtn := widget.NewButton("Add", s.showAddDialog)
	removeBtn := widget.NewButton("Remove", s.removeSelected)
	refreshBtn := widget.NewButton("Refresh", s.refresh)
	changeRepoBtn := widget.NewButton("Change repo", func() {
		s.window.SetContent(s.welcomeView())
	})

	header := container.NewHBox(
		widget.NewLabelWithStyle(s.repoPath, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		changeRepoBtn,
	)

	toolbar := container.NewHBox(addBtn, removeBtn, layout.NewSpacer(), refreshBtn)

	headerCols := ui.NewWorktreeCenter(&s.rowMetrics.DirWidth,
		widget.NewLabelWithStyle("Dir", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Branch", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Path", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	openHeader := container.NewHBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Open", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
	)
	s.colHeader = container.NewBorder(nil, nil, ui.PinColumnSpacer(), openHeader, headerCols)

	listPanel := container.NewBorder(s.colHeader, nil, nil, nil, s.list)

	return container.NewBorder(
		container.NewVBox(header, widget.NewSeparator(), toolbar),
		s.status,
		nil, nil,
		listPanel,
	)
}

func (s *appState) fitColumnsToRow(row fyne.CanvasObject) {
	border, ok := row.(*fyne.Container)
	if !ok {
		return
	}
	center, ok := border.Objects[0].(*fyne.Container)
	if !ok {
		return
	}
	centerWidth := center.Size().Width
	if centerWidth <= 0 {
		return
	}
	pathHalf := ui.FlexHalfWidth(centerWidth, s.rowMetrics.DirWidth)
	truncate := ui.ComputeTruncatePath(s.worktrees, pathHalf)
	next := ui.RowMetrics{DirWidth: s.rowMetrics.DirWidth, TruncatePath: truncate}
	if next == s.rowMetrics {
		return
	}
	s.rowMetrics = next
	s.list.Refresh()
	if s.colHeader != nil {
		s.colHeader.Refresh()
	}
}

func (s *appState) refresh() {
	wts, err := gitops.ListWorktrees(s.repoPath)
	if err != nil {
		s.setStatus("Error: " + err.Error())
		return
	}
	branches, err := gitops.ListBranches(s.repoPath)
	if err != nil {
		s.setStatus("Error: " + err.Error())
		return
	}
	s.worktrees = s.sortWorktrees(wts)
	s.branches = branches
	s.rowMetrics = ui.RowMetrics{DirWidth: ui.ComputeDirWidth(s.worktrees)}
	s.ideAvail = ide.Detect()
	s.selectedID = -1
	s.list.UnselectAll()
	s.list.Refresh()
	if s.colHeader != nil {
		s.colHeader.Refresh()
	}
	s.setStatus(fmt.Sprintf("%d worktree(s)", len(wts)))
}

func (s *appState) setStatus(msg string) {
	if s.status != nil {
		s.status.SetText(msg)
	}
}

func (s *appState) bindIDEButton(btns []*widget.Button, idx int, kind ide.Kind, path string, available bool) {
	if idx >= len(btns) {
		return
	}
	btn := btns[idx]
	if available {
		btn.Enable()
		btn.OnTapped = func() { s.openInIDE(path, kind) }
	} else {
		btn.Disable()
		btn.OnTapped = nil
	}
}

func (s *appState) openInIDE(path string, kind ide.Kind) {
	if kind == ide.Claude && !s.ideAvail.Claude && s.ideAvail.ClaudeDesktopApp {
		dialog.ShowInformation("Claude Code",
			"The Claude desktop app is installed, but Claude Code is not available on your plan.\n\n"+
				"Claude Code in the desktop app requires a Pro or Max subscription.\n\n"+
				"Install the Claude Code CLI to use it from the terminal:\n\n"+
				"curl -fsSL https://claude.ai/install.sh | bash\n\n"+
				"Then click Refresh in this app.",
			s.window)
		return
	}
	if err := ide.Open(path, kind); err != nil {
		dialog.ShowError(err, s.window)
	}
}

func (s *appState) copyToClipboard(text, label string) {
	if text == "" {
		s.setStatus(fmt.Sprintf("Nothing to copy for %s", label))
		return
	}
	s.window.Clipboard().SetContent(text)
	s.setStatus(fmt.Sprintf("Copied %s", label))
}

func (s *appState) pinEntry(wtPath string) string {
	repo := s.normalizedPath(s.repoPath)
	worktree := s.normalizedPath(wtPath)
	return repo + "\x00" + worktree
}

func (s *appState) normalizedPath(path string) string {
	norm, err := gitops.NormalizePath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return norm
}

func (s *appState) pinnedPaths() []string {
	repo := s.normalizedPath(s.repoPath)
	var paths []string
	for _, entry := range s.prefs.StringList(pinnedWorktreesKey) {
		repoPath, wtPath, ok := splitPinEntry(entry)
		if !ok || repoPath != repo {
			continue
		}
		paths = append(paths, s.normalizedPath(wtPath))
	}
	return paths
}

func (s *appState) isPinned(wtPath string) bool {
	normWT := s.normalizedPath(wtPath)
	normRepo := s.normalizedPath(s.repoPath)
	for _, e := range s.prefs.StringList(pinnedWorktreesKey) {
		repo, storedWT, ok := splitPinEntry(e)
		if ok && repo == normRepo && s.normalizedPath(storedWT) == normWT {
			return true
		}
	}
	return false
}

func (s *appState) togglePin(wtPath string) {
	entries := s.prefs.StringList(pinnedWorktreesKey)
	normWT := s.normalizedPath(wtPath)
	normRepo := s.normalizedPath(s.repoPath)
	for i, e := range entries {
		repo, storedWT, ok := splitPinEntry(e)
		if ok && repo == normRepo && s.normalizedPath(storedWT) == normWT {
			entries = append(entries[:i], entries[i+1:]...)
			s.prefs.SetStringList(pinnedWorktreesKey, entries)
			s.refresh()
			s.setStatus("Unpinned worktree")
			return
		}
	}
	if len(s.pinnedPaths()) >= maxPinnedWorktrees {
		dialog.ShowInformation("Pin", fmt.Sprintf("You can pin at most %d worktrees per repository", maxPinnedWorktrees), s.window)
		return
	}
	entries = append(entries, s.pinEntry(wtPath))
	s.prefs.SetStringList(pinnedWorktreesKey, entries)
	s.refresh()
	s.setStatus("Pinned worktree")
}

func (s *appState) sortWorktrees(wts []gitops.Worktree) []gitops.Worktree {
	pinned := s.pinnedPaths()
	pinOrder := make(map[string]int, len(pinned))
	for i, p := range pinned {
		pinOrder[p] = i
	}

	existing := make(map[string]gitops.Worktree, len(wts))
	for _, wt := range wts {
		existing[s.normalizedPath(wt.Path)] = wt
	}

	var sorted []gitops.Worktree
	for _, p := range pinned {
		if wt, ok := existing[p]; ok {
			sorted = append(sorted, wt)
		}
	}
	for _, wt := range wts {
		if _, ok := pinOrder[s.normalizedPath(wt.Path)]; !ok {
			sorted = append(sorted, wt)
		}
	}

	s.prunePinnedPaths(existing)
	return sorted
}

func (s *appState) prunePinnedPaths(existing map[string]gitops.Worktree) {
	repo := s.normalizedPath(s.repoPath)
	var kept []string
	for _, entry := range s.prefs.StringList(pinnedWorktreesKey) {
		repoPath, wtPath, ok := splitPinEntry(entry)
		if !ok {
			continue
		}
		if repoPath != repo {
			kept = append(kept, entry)
			continue
		}
		if _, ok := existing[s.normalizedPath(wtPath)]; ok {
			kept = append(kept, s.pinEntry(wtPath))
		}
	}
	s.prefs.SetStringList(pinnedWorktreesKey, kept)
}

func splitPinEntry(entry string) (repoPath, wtPath string, ok bool) {
	if i := strings.IndexByte(entry, '\x00'); i >= 0 {
		return entry[:i], entry[i+1:], true
	}
	if i := strings.IndexByte(entry, '|'); i >= 0 {
		return entry[:i], entry[i+1:], true
	}
	return "", "", false
}

func filterBranches(branches []string, query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]string, 0, len(branches))
	for _, branch := range branches {
		if query == "" || strings.Contains(strings.ToLower(branch), query) {
			filtered = append(filtered, branch)
		}
	}
	return filtered
}

func preferredBranch(branches []string, preferred string) string {
	for _, branch := range branches {
		if branch == preferred {
			return preferred
		}
	}
	if len(branches) > 0 {
		return branches[0]
	}
	return ""
}

func branchSelectorPanel(branches []string, preferred string) (*widget.Select, fyne.CanvasObject) {
	branchSearch := widget.NewEntry()
	branchSearch.SetPlaceHolder("Search branches...")

	branchSelect := widget.NewSelect(filterBranches(branches, ""), nil)
	if selected := preferredBranch(branches, preferred); selected != "" {
		branchSelect.SetSelected(selected)
	}

	branchSearch.OnChanged = func(query string) {
		filtered := filterBranches(branches, query)
		current := branchSelect.Selected
		branchSelect.SetOptions(filtered)
		if len(filtered) == 0 {
			branchSelect.ClearSelected()
			return
		}
		for _, branch := range filtered {
			if branch == current {
				branchSelect.SetSelected(current)
				return
			}
		}
		branchSelect.SetSelected(filtered[0])
	}

	return branchSelect, container.NewVBox(branchSearch, branchSelect)
}

func (s *appState) showAddDialog() {
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("/path/to/new/worktree")

	modeSelect := widget.NewSelect([]string{"Existing branch", "New branch"}, nil)
	modeSelect.SetSelected("Existing branch")

	sortedBranches := append([]string(nil), s.branches...)
	sort.Strings(sortedBranches)

	branchSelect, existingBranchPanel := branchSelectorPanel(sortedBranches, "")
	sourceBranchSelect, sourceBranchPanel := branchSelectorPanel(sortedBranches, "develop")

	newBranchEntry := widget.NewEntry()
	newBranchEntry.SetPlaceHolder("feature/my-branch")

	newBranchPanel := container.NewVBox(
		widget.NewLabel("New branch name"),
		newBranchEntry,
		widget.NewLabel("Source branch"),
		sourceBranchPanel,
	)
	newBranchPanel.Hide()

	modeSelect.OnChanged = func(sel string) {
		if sel == "New branch" {
			existingBranchPanel.Hide()
			newBranchPanel.Show()
		} else {
			newBranchPanel.Hide()
			existingBranchPanel.Show()
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Path", pathEntry),
		widget.NewFormItem("Mode", modeSelect),
		widget.NewFormItem("Branch", container.NewStack(existingBranchPanel, newBranchPanel)),
	)

	d := dialog.NewCustomConfirm("Add worktree", "Create", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		wtPath := strings.TrimSpace(pathEntry.Text)
		if wtPath == "" {
			dialog.ShowError(fmt.Errorf("worktree path is required"), s.window)
			return
		}
		abs, err := filepath.Abs(wtPath)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}

		newBranch := modeSelect.Selected == "New branch"
		var branch, sourceBranch string
		if newBranch {
			branch = strings.TrimSpace(newBranchEntry.Text)
			sourceBranch = sourceBranchSelect.Selected
			if branch == "" {
				dialog.ShowError(fmt.Errorf("branch name is required"), s.window)
				return
			}
		} else {
			branch = branchSelect.Selected
			if branch == "" {
				dialog.ShowError(fmt.Errorf("select a branch"), s.window)
				return
			}
		}

		if err := gitops.AddWorktree(s.repoPath, abs, branch, newBranch, sourceBranch); err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		s.refresh()
	}, s.window)
	d.Resize(fyne.NewSize(420, 320))
	d.Show()
}

func (s *appState) removeSelected() {
	id := s.selectedID
	if id < 0 || int(id) >= len(s.worktrees) {
		dialog.ShowInformation("Remove", "Select a worktree first", s.window)
		return
	}
	wt := s.worktrees[id]
	if wt.Main {
		dialog.ShowInformation("Remove", "Cannot remove the main worktree", s.window)
		return
	}

	branch := wt.Branch
	if branch == "" {
		branch = "(detached)"
	}

	msg := fmt.Sprintf("Remove this worktree?\n\nBranch: %s\nPath: %s", branch, wt.Path)
	dialog.ShowConfirm("Remove worktree", msg, func(ok bool) {
		if !ok {
			return
		}
		if err := gitops.RemoveWorktree(s.repoPath, wt.Path, false); err != nil {
			dialog.ShowConfirm("Force remove?", err.Error()+"\n\nForce remove?", func(force bool) {
				if !force {
					return
				}
				if err := gitops.RemoveWorktree(s.repoPath, wt.Path, true); err != nil {
					dialog.ShowError(err, s.window)
					return
				}
				s.refresh()
			}, s.window)
			return
		}
		s.refresh()
	}, s.window)
}
