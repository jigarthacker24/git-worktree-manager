package ui

import (
	"fmt"
	"strings"

	"github.com/jigarthacker24/git-worktree-manager/internal/tasks"

	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const taskPanelMinHeight = float32(180)

// TaskPanel is a global todo list with tag filtering.
type TaskPanel struct {
	widget.BaseWidget
	store         *tasks.Store
	window        fyne.Window
	filterTag     string
	allTasks      []tasks.Task
	visibleTasks  []tasks.Task
	list          *widget.List
	filterBox     *fyne.Container
	status        *widget.Label
	nameEntry     *widget.Entry
	tagEntry      *widget.Entry
	onChange      func()
}

// NewTaskPanel creates a persistent bottom task panel.
func NewTaskPanel(store *tasks.Store, window fyne.Window, onChange func()) *TaskPanel {
	p := &TaskPanel{store: store, window: window, onChange: onChange}
	p.ExtendBaseWidget(p)
	p.reload()
	return p
}

func (p *TaskPanel) reload() {
	p.loadTasks()
	p.refreshList()
	p.rebuildFilterChips()
}

func (p *TaskPanel) loadTasks() {
	p.allTasks = p.store.List()
	p.visibleTasks = tasks.FilterByTag(p.allTasks, p.filterTag)
}

func (p *TaskPanel) refreshList() {
	if p.list != nil {
		p.list.Refresh()
	}
	if p.status != nil {
		p.status.SetText(p.statusText())
	}
}

func (p *TaskPanel) setTaskCompleted(taskID string, completed bool) {
	if err := p.store.SetCompleted(taskID, completed); err != nil {
		return
	}
	fyne.Do(func() {
		p.loadTasks()
		p.refreshList()
	})
}

func (p *TaskPanel) moveTask(taskID string, delta int) {
	activeIDs := tasks.ActiveIDsInOrder(tasks.ActiveTasks(p.visibleTasks))
	if err := p.store.MoveActiveAmong(activeIDs, taskID, delta); err != nil {
		return
	}
	fyne.Do(func() {
		p.loadTasks()
		p.refreshList()
	})
}

func bindMoveButton(btn *ttwidget.Button, enabled bool, tooltip string, onTap func()) {
	if btn == nil {
		return
	}
	SetToolTip(btn, tooltip)
	if enabled {
		btn.Enable()
		btn.OnTapped = onTap
	} else {
		btn.Disable()
		btn.OnTapped = nil
	}
}

func applyTaskNameLabel(lbl *widget.Label, name string, completed bool) {
	if completed {
		lbl.SetText(strikeThrough(name))
		lbl.TextStyle = fyne.TextStyle{Italic: true}
	} else {
		lbl.SetText(name)
		lbl.TextStyle = fyne.TextStyle{}
	}
	lbl.Refresh()
}

func strikeThrough(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		b.WriteRune(r)
		b.WriteRune('\u0336')
	}
	return b.String()
}

func (p *TaskPanel) statusText() string {
	active := len(tasks.ActiveTasks(p.visibleTasks))
	done := len(tasks.CompletedTasks(p.visibleTasks))
	total := len(p.visibleTasks)

	var summary string
	switch {
	case done == 0:
		summary = fmtTaskCount(total)
	case active == 0:
		summary = fmt.Sprintf("%d completed", done)
	default:
		summary = fmt.Sprintf("%d active, %d completed", active, done)
	}
	if p.filterTag != "" {
		return summary + " (tag: " + p.filterTag + ")"
	}
	return summary
}

func fmtTaskCount(n int) string {
	if n == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", n)
}

func (p *TaskPanel) setFilter(tag string) {
	p.filterTag = tasks.NormalizeTag(tag)
	p.reload()
}

func (p *TaskPanel) rebuildFilterChips() {
	if p.filterBox == nil {
		return
	}
	objs := []fyne.CanvasObject{p.filterChip("All", p.filterTag == "")}
	for _, tag := range tasks.UniqueTags(p.allTasks) {
		objs = append(objs, p.filterChip(tag, p.filterTag == tag))
	}
	p.filterBox.Objects = objs
	p.filterBox.Refresh()
}

func (p *TaskPanel) filterChip(label string, selected bool) fyne.CanvasObject {
	tag := label
	if label == "All" {
		tag = ""
	}
	btn := widget.NewButton(label, func() {
		p.setFilter(tag)
	})
	if selected {
		btn.Importance = widget.HighImportance
	} else {
		btn.Importance = widget.LowImportance
	}
	return btn
}

func (p *TaskPanel) addTask() {
	name := p.nameEntry.Text
	if strings.TrimSpace(name) == "" {
		return
	}
	p.store.Add(name, p.tagEntry.Text)
	p.nameEntry.SetText("")
	p.tagEntry.SetText("")
	p.reload()
	if p.onChange != nil {
		p.onChange()
	}
}

func (p *TaskPanel) showEditTaskDialog(t tasks.Task) {
	if p.window == nil {
		return
	}
	nameEntry := widget.NewEntry()
	nameEntry.SetText(t.Name)
	nameEntry.SetPlaceHolder("Task name")

	tagsEntry := widget.NewEntry()
	tagsEntry.SetText(tasks.TagsInputString(t.Tags))
	tagsEntry.SetPlaceHolder("Tags (comma-separated)")

	content := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Tags", tagsEntry),
	)

	d := dialog.NewCustomConfirm("Edit task", "Save", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			dialog.ShowError(fmt.Errorf("task name is required"), p.window)
			return
		}
		if err := p.store.UpdateTask(t.ID, name, tagsEntry.Text); err != nil {
			dialog.ShowError(err, p.window)
			return
		}
		fyne.Do(func() {
			p.reload()
		})
	}, p.window)
	d.Resize(fyne.NewSize(420, 180))
	d.Show()
}

func (p *TaskPanel) activePosition(taskID string) (pos, count int) {
	count = 0
	pos = -1
	for _, t := range p.visibleTasks {
		if t.Completed {
			continue
		}
		if t.ID == taskID {
			pos = count
		}
		count++
	}
	return pos, count
}

func (p *TaskPanel) bindRow(id widget.ListItemID, obj fyne.CanvasObject) {
	if id < 0 || int(id) >= len(p.visibleTasks) {
		return
	}
	t := p.visibleTasks[id]
	row := obj.(*TaskRow)
	activePos, activeCount := p.activePosition(t.ID)
	row.Bind(t, activePos, activeCount)
}

func (p *TaskPanel) CreateRenderer() fyne.WidgetRenderer {
	p.nameEntry = widget.NewEntry()
	p.nameEntry.SetPlaceHolder("New task")

	p.tagEntry = widget.NewEntry()
	p.tagEntry.SetPlaceHolder("Tags (optional, comma-separated)")

	addBtn := NewToolButton("Add", "Add task", p.addTask)
	addBtn.Importance = widget.HighImportance

	p.filterBox = container.NewHBox()
	p.rebuildFilterChips()

	p.list = widget.NewList(
		func() int { return len(p.visibleTasks) },
		func() fyne.CanvasObject { return newTaskRow(p) },
		p.bindRow,
	)

	p.status = widget.NewLabel(p.statusText())

	header := container.NewHBox(
		widget.NewLabelWithStyle("Tasks", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		p.status,
	)

	addRow := container.NewBorder(nil, nil, nil, addBtn,
		container.NewGridWithColumns(2, p.nameEntry, p.tagEntry),
	)

	content := container.NewBorder(
		container.NewVBox(header, p.filterBox, widget.NewSeparator()),
		addRow,
		nil, nil,
		container.NewScroll(p.list),
	)

	r := &taskPanelRenderer{
		panel:   p,
		content: content,
	}
	r.content.Resize(fyne.NewSize(content.MinSize().Width, taskPanelMinHeight))
	return r
}

func (p *TaskPanel) MinSize() fyne.Size {
	return fyne.NewSize(400, taskPanelMinHeight)
}

type taskPanelRenderer struct {
	panel   *TaskPanel
	content *fyne.Container
}

func (r *taskPanelRenderer) Layout(size fyne.Size) {
	h := size.Height
	if h < taskPanelMinHeight {
		h = taskPanelMinHeight
	}
	r.content.Resize(fyne.NewSize(size.Width, h))
	r.content.Move(fyne.NewPos(0, 0))
}

func (r *taskPanelRenderer) MinSize() fyne.Size {
	min := r.content.MinSize()
	if min.Height < taskPanelMinHeight {
		min.Height = taskPanelMinHeight
	}
	return min
}

func (r *taskPanelRenderer) Refresh() {
	r.content.Refresh()
}

func (r *taskPanelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

func (r *taskPanelRenderer) Destroy() {}
