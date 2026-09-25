package ui

import (
	"github.com/jigarthacker24/git-worktree-manager/internal/tasks"

	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type taskCheckBinding struct {
	taskID string
	ignore bool
}

// TaskRow is one task line in the active or completed list.
type TaskRow struct {
	widget.BaseWidget
	panel       *TaskPanel
	bind        taskCheckBinding
	check       *widget.Check
	nameLbl     *widget.Label
	tagLbl      *widget.Label
	editTaskBtn *ttwidget.Button
	upBtn       *ttwidget.Button
	downBtn     *ttwidget.Button
	delBtn      *ttwidget.Button
	content     *fyne.Container
}

func newTaskRow(panel *TaskPanel) *TaskRow {
	r := &TaskRow{panel: panel}
	r.check = widget.NewCheck("", nil)
	r.nameLbl = widget.NewLabel("")
	r.tagLbl = widget.NewLabel("")
	r.tagLbl.Truncation = fyne.TextTruncateEllipsis
	r.editTaskBtn = NewIconToolButton(theme.DocumentCreateIcon(), "Edit task", nil)
	r.upBtn = NewToolButton("↑", "Move up", nil)
	r.downBtn = NewToolButton("↓", "Move down", nil)
	r.delBtn = NewIconToolButton(theme.DeleteIcon(), "Delete task", nil)

	tagCell := container.NewBorder(nil, nil, nil, r.editTaskBtn, r.tagLbl)
	r.content = container.NewBorder(nil, nil, r.check,
		container.NewHBox(r.upBtn, r.downBtn, r.delBtn),
		container.NewGridWithColumns(2, r.nameLbl, tagCell),
	)
	r.ExtendBaseWidget(r)
	return r
}

func (r *TaskRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *TaskRow) Bind(t tasks.Task, activePos, activeCount int) {
	taskID := t.ID
	clearCheckHighlight(r.check)

	r.bind.ignore = true
	r.bind.taskID = taskID
	r.check.OnChanged = nil
	r.check.SetChecked(t.Completed)
	r.bind.ignore = false
	r.check.OnChanged = func(checked bool) {
		if r.bind.ignore || r.bind.taskID != taskID {
			return
		}
		clearCheckHighlight(r.check)
		r.panel.onTaskCheckChanged(taskID, checked)
	}

	applyTaskNameLabel(r.nameLbl, t.Name, t.Completed)

	r.tagLbl.SetText(tasks.TagsLabel(t.Tags))
	r.tagLbl.TextStyle = fyne.TextStyle{}
	r.tagLbl.Refresh()

	if r.editTaskBtn != nil {
		SetToolTip(r.editTaskBtn, "Edit task")
		r.editTaskBtn.OnTapped = func() {
			r.panel.showEditTaskDialog(t)
		}
	}

	canMove := !t.Completed && activeCount > 0 && activePos >= 0
	bindMoveButton(r.upBtn, canMove && activePos > 0, "Move up", func() {
		r.panel.moveTask(taskID, -1)
	})
	bindMoveButton(r.downBtn, canMove && activePos < activeCount-1, "Move down", func() {
		r.panel.moveTask(taskID, 1)
	})
	if r.delBtn != nil {
		r.delBtn.OnTapped = func() {
			_ = r.panel.store.Delete(taskID)
			fyne.Do(func() {
				r.panel.reload()
				if r.panel.onChange != nil {
					r.panel.onChange()
				}
			})
		}
	}
}

func (p *TaskPanel) onTaskCheckChanged(taskID string, completed bool) {
	for _, t := range p.allTasks {
		if t.ID == taskID {
			if t.Completed == completed {
				return
			}
			break
		}
	}
	p.setTaskCompleted(taskID, completed)
}

// clearCheckHighlight removes focus/hover rings left on recycled list checkboxes.
func clearCheckHighlight(check *widget.Check) {
	if check == nil {
		return
	}
	check.FocusLost()
	check.MouseOut()
	if app := fyne.CurrentApp(); app != nil {
		if c := app.Driver().CanvasForObject(check); c != nil && c.Focused() == check {
			c.Unfocus()
		}
	}
}
