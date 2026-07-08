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
	status     *widget.Label
	selectedID widget.ListItemID
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
		recentList := widget.NewList(
			func() int { return len(recentPaths) },
			func() fyne.CanvasObject {
				return widget.NewLabel("template")
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(recentPaths[id])
			},
		)
		recentList.OnSelected = func(id widget.ListItemID) {
			s.openRepo(recentPaths[id])
		}
		content.Add(widget.NewSeparator())
		content.Add(widget.NewLabel("Recent"))
		content.Add(recentList)
	}

	return container.NewBorder(nil, nil, nil, nil, content)
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

			openBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), nil)
			openBtn.Importance = widget.LowImportance
			openCell := ui.WrapWithHint(openBtn, "Open with Cursor", s.setStatus)

			cols := container.NewGridWithColumns(3,
				container.NewHBox(widget.NewLabel("branch"), copyIcon()),
				widget.NewLabel("dir"),
				container.NewHBox(widget.NewLabel("path"), copyIcon()),
			)
			return container.NewBorder(nil, nil, pinBtn, openCell, cols)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			wt := s.worktrees[id]
			border := obj.(*fyne.Container)
			cols := border.Objects[0].(*fyne.Container)
			pinBtn := border.Objects[1].(*widget.Button)
			openBtn := ui.ButtonFromHint(border.Objects[2])

			branchBox := cols.Objects[0].(*fyne.Container)
			dirLbl := cols.Objects[1].(*widget.Label)
			pathBox := cols.Objects[2].(*fyne.Container)

			branchLbl := branchBox.Objects[0].(*widget.Label)
			copyBranchBtn := branchBox.Objects[1].(*widget.Button)
			pathLbl := pathBox.Objects[0].(*widget.Label)
			copyPathBtn := pathBox.Objects[1].(*widget.Button)

			branch := wt.Branch
			if branch == "" {
				branch = "(detached)"
			}
			displayBranch := branch
			if wt.Main {
				displayBranch += " · main"
			}
			branchLbl.SetText(displayBranch)
			dirLbl.SetText(wt.DirName)
			pathLbl.SetText(wt.Path)

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
			openBtn.OnTapped = func() {
				s.openInCursor(wt.Path)
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

	headerCols := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Branch", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Dir", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Path", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	openHeader := container.NewCenter(
		widget.NewLabelWithStyle("Open", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	columnHeader := container.NewBorder(nil, nil, widget.NewLabel(""), openHeader, headerCols)

	listPanel := container.NewBorder(columnHeader, nil, nil, nil, s.list)

	return container.NewBorder(
		container.NewVBox(header, widget.NewSeparator(), toolbar),
		s.status,
		nil, nil,
		listPanel,
	)
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
	s.selectedID = -1
	s.list.UnselectAll()
	s.list.Refresh()
	s.setStatus(fmt.Sprintf("%d worktree(s)", len(wts)))
}

func (s *appState) setStatus(msg string) {
	if s.status != nil {
		s.status.SetText(msg)
	}
}

func (s *appState) openInCursor(path string) {
	if err := ide.OpenInCursor(path); err != nil {
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

func (s *appState) showAddDialog() {
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("/path/to/new/worktree")

	modeSelect := widget.NewSelect([]string{"Existing branch", "New branch"}, nil)
	modeSelect.SetSelected("Existing branch")

	sortedBranches := append([]string(nil), s.branches...)
	sort.Strings(sortedBranches)

	branchSearch := widget.NewEntry()
	branchSearch.SetPlaceHolder("Search branches...")

	branchSelect := widget.NewSelect(filterBranches(sortedBranches, ""), nil)
	if len(sortedBranches) > 0 {
		branchSelect.SetSelected(sortedBranches[0])
	}

	branchSearch.OnChanged = func(query string) {
		filtered := filterBranches(sortedBranches, query)
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

	existingBranchPanel := container.NewVBox(branchSearch, branchSelect)

	newBranchEntry := widget.NewEntry()
	newBranchEntry.SetPlaceHolder("feature/my-branch")
	newBranchEntry.Hide()

	modeSelect.OnChanged = func(sel string) {
		if sel == "New branch" {
			existingBranchPanel.Hide()
			newBranchEntry.Show()
		} else {
			newBranchEntry.Hide()
			existingBranchPanel.Show()
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Path", pathEntry),
		widget.NewFormItem("Mode", modeSelect),
		widget.NewFormItem("Branch", container.NewStack(existingBranchPanel, newBranchEntry)),
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
		branch := branchSelect.Selected
		if newBranch {
			branch = strings.TrimSpace(newBranchEntry.Text)
			if branch == "" {
				dialog.ShowError(fmt.Errorf("branch name is required"), s.window)
				return
			}
		} else if branch == "" {
			dialog.ShowError(fmt.Errorf("select a branch"), s.window)
			return
		}

		if err := gitops.AddWorktree(s.repoPath, abs, branch, newBranch); err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		s.refresh()
	}, s.window)
	d.Resize(fyne.NewSize(420, 260))
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
