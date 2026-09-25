package tasks

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const PrefKey = "global_tasks"

// Task is a single global todo item.
type Task struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Completed bool     `json:"completed"`
}

// Store persists tasks via load/save callbacks (typically Fyne preferences).
type Store struct {
	load func() string
	save func(string)
}

// NewStore creates a task store backed by the given callbacks.
func NewStore(load func() string, save func(string)) *Store {
	return &Store{load: load, save: save}
}

// List returns all stored tasks with active tasks first and completed at the bottom.
func (s *Store) List() []Task {
	return DisplayOrder(ParseJSON(s.load()))
}

// Save replaces all tasks.
func (s *Store) Save(all []Task) {
	s.save(MarshalJSON(all))
}

// Add appends a new active task and returns it.
// tagsInput is comma-separated (e.g. "frontend, review").
func (s *Store) Add(name, tagsInput string) Task {
	name = strings.TrimSpace(name)
	tags := ParseTagsInput(tagsInput)
	all := ParseJSON(s.load())
	t := Task{
		ID:   newID(),
		Name: name,
		Tags: tags,
	}
	var active, done []Task
	for _, item := range all {
		if item.Completed {
			done = append(done, item)
		} else {
			active = append(active, item)
		}
	}
	active = append(active, t)
	s.Save(append(active, done...))
	return t
}

// SetCompleted updates the completed flag and moves completed tasks to the bottom.
func (s *Store) SetCompleted(id string, completed bool) error {
	all := ParseJSON(s.load())
	var task Task
	found := false
	rest := make([]Task, 0, len(all))
	for _, t := range all {
		if t.ID != id {
			rest = append(rest, t)
			continue
		}
		task = t
		task.Completed = completed
		found = true
	}
	if !found {
		return fmt.Errorf("task not found: %s", id)
	}

	var active, done []Task
	for _, t := range rest {
		if t.Completed {
			done = append(done, t)
		} else {
			active = append(active, t)
		}
	}
	if completed {
		done = append(done, task)
	} else {
		active = append(active, task)
	}
	s.Save(append(active, done...))
	return nil
}

// Delete removes a task by ID.
func (s *Store) Delete(id string) error {
	all := s.List()
	out := all[:0]
	found := false
	for _, t := range all {
		if t.ID == id {
			found = true
			continue
		}
		out = append(out, t)
	}
	if !found {
		return fmt.Errorf("task not found: %s", id)
	}
	s.Save(out)
	return nil
}

// UpdateTask changes the task name and tags. tagsInput is comma-separated.
func (s *Store) UpdateTask(id, name, tagsInput string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("task name is required")
	}
	all := ParseJSON(s.load())
	found := false
	tags := ParseTagsInput(tagsInput)
	for i := range all {
		if all[i].ID != id {
			continue
		}
		all[i].Name = name
		all[i].Tags = tags
		found = true
		break
	}
	if !found {
		return fmt.Errorf("task not found: %s", id)
	}
	s.Save(DisplayOrder(all))
	return nil
}

// MoveActiveAmong moves an active task up or down within orderedVisibleActiveIDs.
// delta is -1 for up and +1 for down.
func (s *Store) MoveActiveAmong(orderedVisibleActiveIDs []string, id string, delta int) error {
	if delta != -1 && delta != 1 {
		return fmt.Errorf("invalid move delta: %d", delta)
	}
	idx := -1
	for i, visibleID := range orderedVisibleActiveIDs {
		if visibleID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("task not found in visible active list: %s", id)
	}
	next := idx + delta
	if next < 0 || next >= len(orderedVisibleActiveIDs) {
		return nil
	}

	all := ParseJSON(s.load())
	if err := swapTasksByID(all, id, orderedVisibleActiveIDs[next]); err != nil {
		return err
	}
	s.Save(DisplayOrder(all))
	return nil
}

// DisplayOrder returns active tasks first, then completed, preserving order within each group.
func DisplayOrder(all []Task) []Task {
	if len(all) == 0 {
		return nil
	}
	var active, done []Task
	for _, t := range all {
		if t.Completed {
			done = append(done, t)
		} else {
			active = append(active, t)
		}
	}
	return append(active, done...)
}

// ActiveTasks returns incomplete tasks preserving order.
func ActiveTasks(all []Task) []Task {
	var out []Task
	for _, t := range all {
		if !t.Completed {
			out = append(out, t)
		}
	}
	return out
}

// CompletedTasks returns completed tasks preserving order.
func CompletedTasks(all []Task) []Task {
	var out []Task
	for _, t := range all {
		if t.Completed {
			out = append(out, t)
		}
	}
	return out
}

// TaskByID finds a task by ID.
func TaskByID(all []Task, id string) (Task, bool) {
	for _, t := range all {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// ActiveIDsInOrder returns IDs of active tasks from list in order.
func ActiveIDsInOrder(all []Task) []string {
	var ids []string
	for _, t := range all {
		if !t.Completed {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

func swapTasksByID(all []Task, idA, idB string) error {
	iA, iB := -1, -1
	for i := range all {
		switch all[i].ID {
		case idA:
			iA = i
		case idB:
			iB = i
		}
	}
	if iA < 0 || iB < 0 {
		return fmt.Errorf("task not found for swap")
	}
	all[iA], all[iB] = all[iB], all[iA]
	return nil
}

// NormalizeTag trims whitespace from a tag label.
func NormalizeTag(tag string) string {
	return strings.TrimSpace(tag)
}

// ParseTagsInput splits comma-separated tag text into normalized unique tags.
func ParseTagsInput(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	return NormalizeTagsList(strings.Split(input, ","))
}

// NormalizeTagsList trims, deduplicates, and drops empty tags preserving order.
func NormalizeTagsList(tags []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = NormalizeTag(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

// HasTag reports whether a task includes the given tag.
func HasTag(t Task, tag string) bool {
	tag = NormalizeTag(tag)
	if tag == "" {
		return false
	}
	for _, item := range t.Tags {
		if NormalizeTag(item) == tag {
			return true
		}
	}
	return false
}

// TagsInputString formats tags for editing in a comma-separated field.
func TagsInputString(tags []string) string {
	return strings.Join(NormalizeTagsList(tags), ", ")
}

// TagsLabel formats tags for display.
func TagsLabel(tags []string) string {
	tags = NormalizeTagsList(tags)
	if len(tags) == 0 {
		return "—"
	}
	return strings.Join(tags, ", ")
}

// ParseJSON decodes tasks from stored JSON; invalid or empty input yields nil slice.
func ParseJSON(raw string) []Task {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var rawTasks []struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Tag       string   `json:"tag"`
		Tags      []string `json:"tags"`
		Completed bool     `json:"completed"`
	}
	if err := json.Unmarshal([]byte(raw), &rawTasks); err != nil {
		return nil
	}
	out := make([]Task, 0, len(rawTasks))
	for _, rt := range rawTasks {
		tags := NormalizeTagsList(rt.Tags)
		if len(tags) == 0 && rt.Tag != "" {
			tags = []string{NormalizeTag(rt.Tag)}
		}
		out = append(out, Task{
			ID:        rt.ID,
			Name:      rt.Name,
			Tags:      tags,
			Completed: rt.Completed,
		})
	}
	return out
}

// MarshalJSON encodes tasks for storage.
func MarshalJSON(all []Task) string {
	if len(all) == 0 {
		return "[]"
	}
	b, err := json.Marshal(all)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// FilterByTag returns tasks that include the tag; empty tag returns all tasks.
func FilterByTag(all []Task, tag string) []Task {
	tag = NormalizeTag(tag)
	if tag == "" {
		return append([]Task(nil), all...)
	}
	var out []Task
	for _, t := range all {
		if HasTag(t, tag) {
			out = append(out, t)
		}
	}
	return out
}

// UniqueTags returns sorted distinct non-empty tags from all tasks.
func UniqueTags(all []Task) []string {
	seen := make(map[string]struct{})
	for _, t := range all {
		for _, tag := range NormalizeTagsList(t.Tags) {
			seen[tag] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for tag := range seen {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func newID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
