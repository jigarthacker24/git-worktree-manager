package tasks

import (
	"testing"
)

func memStore() (*Store, *string) {
	var raw string
	s := NewStore(
		func() string { return raw },
		func(v string) { raw = v },
	)
	return s, &raw
}

func TestParseJSONEmpty(t *testing.T) {
	if got := ParseJSON(""); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := ParseJSON("not json"); got != nil {
		t.Fatalf("expected nil for invalid json, got %v", got)
	}
}

func TestParseJSONLegacySingleTag(t *testing.T) {
	raw := `[{"id":"1","name":"x","tag":"legacy","completed":false}]`
	got := ParseJSON(raw)
	if len(got) != 1 || len(got[0].Tags) != 1 || got[0].Tags[0] != "legacy" {
		t.Fatalf("expected legacy tag migration, got %+v", got)
	}
}

func TestParseTagsInput(t *testing.T) {
	got := ParseTagsInput(" frontend, review , frontend , ")
	if len(got) != 2 || got[0] != "frontend" || got[1] != "review" {
		t.Fatalf("unexpected tags: %v", got)
	}
	if ParseTagsInput("") != nil {
		t.Fatal("empty input should return nil")
	}
}

func TestHasTag(t *testing.T) {
	t1 := Task{Tags: []string{"a", "b"}}
	if !HasTag(t1, "a") || !HasTag(t1, "b") || HasTag(t1, "c") {
		t.Fatal("HasTag mismatch")
	}
}

func TestTagsLabel(t *testing.T) {
	if TagsLabel(nil) != "—" {
		t.Fatal("expected em dash for empty tags")
	}
	if TagsLabel([]string{"x", "y"}) != "x, y" {
		t.Fatal("expected comma-separated label")
	}
}

func TestStoreAddAndList(t *testing.T) {
	s, raw := memStore()
	t1 := s.Add("Fix tests", "ci, refactor")
	t2 := s.Add("Review PR", "")

	all := s.List()
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(all))
	}
	if t1.Name != "Fix tests" || len(t1.Tags) != 2 || t1.Tags[0] != "ci" || t1.Completed {
		t.Fatalf("unexpected first task: %+v", t1)
	}
	if len(t2.Tags) != 0 {
		t.Fatalf("expected empty tags, got %v", t2.Tags)
	}
	if *raw == "" {
		t.Fatal("expected persisted JSON")
	}
}

func TestStoreSetCompleted(t *testing.T) {
	s, _ := memStore()
	t1 := s.Add("First", "work")
	t2 := s.Add("Second", "work")
	if err := s.SetCompleted(t1.ID, true); err != nil {
		t.Fatal(err)
	}
	all := s.List()
	if len(all) != 2 || !all[1].Completed || all[1].ID != t1.ID || all[0].ID != t2.ID {
		t.Fatalf("completed task should move to bottom: %+v", all)
	}
	if err := s.SetCompleted(t1.ID, false); err != nil {
		t.Fatal(err)
	}
	all = s.List()
	if len(all) != 2 || all[0].ID != t2.ID || all[1].ID != t1.ID || all[1].Completed {
		t.Fatalf("reactivated task should join end of active section: %+v", all)
	}
	if err := s.SetCompleted("missing", true); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestTwoTasksDifferentTagsCompleteFirstOnly(t *testing.T) {
	s, _ := memStore()
	a := s.Add("Task A", "frontend")
	b := s.Add("Task B", "backend")

	if err := s.SetCompleted(a.ID, true); err != nil {
		t.Fatal(err)
	}

	all := s.List()
	gotB, ok := TaskByID(all, b.ID)
	if !ok || gotB.Completed {
		t.Fatalf("second task must stay active: %+v", gotB)
	}
	gotA, ok := TaskByID(all, a.ID)
	if !ok || !gotA.Completed {
		t.Fatalf("first task must be completed: %+v", gotA)
	}
	if all[0].ID != b.ID || all[1].ID != a.ID {
		t.Fatalf("expected active task first then completed: %+v", all)
	}

	filtered := FilterByTag(all, "backend")
	if len(filtered) != 1 || filtered[0].ID != b.ID || filtered[0].Completed {
		t.Fatalf("backend filter should show only active B: %+v", filtered)
	}
}

func TestTagsInputString(t *testing.T) {
	got := TagsInputString([]string{"b", "a", "a"})
	if got != "b, a" {
		t.Fatalf("unexpected tags input string: %q", got)
	}
}

func TestStoreUpdateTask(t *testing.T) {
	s, _ := memStore()
	t1 := s.Add("Old name", "old")
	if err := s.UpdateTask(t1.ID, "New name", "new, urgent"); err != nil {
		t.Fatal(err)
	}
	got, ok := TaskByID(s.List(), t1.ID)
	if !ok || got.Name != "New name" || len(got.Tags) != 2 || got.Tags[0] != "new" || got.Tags[1] != "urgent" {
		t.Fatalf("unexpected task after edit: %+v", got)
	}
	if err := s.UpdateTask("missing", "x", "y"); err == nil {
		t.Fatal("expected error for missing task")
	}
	if err := s.UpdateTask(t1.ID, "  ", "x"); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStoreUpdateTaskClearsTags(t *testing.T) {
	s, _ := memStore()
	t1 := s.Add("Task", "a, b")
	if err := s.UpdateTask(t1.ID, "Task", ""); err != nil {
		t.Fatal(err)
	}
	got, _ := TaskByID(s.List(), t1.ID)
	if len(got.Tags) != 0 {
		t.Fatalf("expected empty tags, got %v", got.Tags)
	}
}

func TestFilterByTagMatchesAnyTag(t *testing.T) {
	s, _ := memStore()
	t1 := s.Add("Multi", "frontend, backend")
	if len(FilterByTag(s.List(), "frontend")) != 1 {
		t.Fatal("should match first tag")
	}
	if len(FilterByTag(s.List(), "backend")) != 1 {
		t.Fatal("should match second tag")
	}
	if len(FilterByTag(s.List(), "missing")) != 0 {
		t.Fatal("should not match unrelated tag")
	}
	_ = t1
}

func TestStoreDelete(t *testing.T) {
	s, _ := memStore()
	t1 := s.Add("One", "")
	s.Add("Two", "")
	if err := s.Delete(t1.ID); err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 1 {
		t.Fatalf("expected 1 task after delete, got %d", len(s.List()))
	}
	if err := s.Delete("missing"); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestFilterByTag(t *testing.T) {
	all := []Task{
		{Name: "a", Tags: []string{"bug"}},
		{Name: "b", Tags: []string{"feat"}},
		{Name: "c", Tags: []string{"bug", "urgent"}},
	}
	filtered := FilterByTag(all, "bug")
	if len(filtered) != 2 {
		t.Fatalf("expected 2, got %d", len(filtered))
	}
	if len(FilterByTag(all, "")) != 3 {
		t.Fatal("empty filter should return all")
	}
}

func TestActiveAndCompletedTasks(t *testing.T) {
	all := []Task{
		{ID: "1", Name: "a", Completed: true},
		{ID: "2", Name: "b"},
		{ID: "3", Name: "c", Completed: true},
		{ID: "4", Name: "d"},
	}
	active := ActiveTasks(all)
	if len(active) != 2 || active[0].ID != "2" || active[1].ID != "4" {
		t.Fatalf("unexpected active tasks: %+v", active)
	}
	done := CompletedTasks(all)
	if len(done) != 2 || done[0].ID != "1" || done[1].ID != "3" {
		t.Fatalf("unexpected completed tasks: %+v", done)
	}
}

func TestUniqueTags(t *testing.T) {
	all := []Task{
		{Tags: []string{"z"}},
		{Tags: []string{"a", "b"}},
		{Tags: []string{"a"}},
		{Tags: nil},
	}
	got := UniqueTags(all)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "z" {
		t.Fatalf("unexpected tags: %v", got)
	}
}

func TestNormalizeTag(t *testing.T) {
	if NormalizeTag("  hello  ") != "hello" {
		t.Fatal("expected trimmed tag")
	}
}

func TestMarshalJSONRoundTrip(t *testing.T) {
	all := []Task{{ID: "1", Name: "x", Tags: []string{"t", "u"}, Completed: true}}
	raw := MarshalJSON(all)
	parsed := ParseJSON(raw)
	if len(parsed) != 1 || !parsed[0].Completed || len(parsed[0].Tags) != 2 {
		t.Fatalf("round trip failed: %+v", parsed)
	}
}

func TestDisplayOrder(t *testing.T) {
	all := []Task{
		{ID: "1", Name: "done", Completed: true},
		{ID: "2", Name: "active"},
		{ID: "3", Name: "done2", Completed: true},
	}
	ordered := DisplayOrder(all)
	if ordered[0].ID != "2" || ordered[1].ID != "1" || ordered[2].ID != "3" {
		t.Fatalf("unexpected order: %+v", ordered)
	}
}

func TestMoveActiveAmong(t *testing.T) {
	s, _ := memStore()
	a := s.Add("A", "")
	b := s.Add("B", "")
	c := s.Add("C", "")

	if err := s.MoveActiveAmong(ActiveIDsInOrder(s.List()), b.ID, -1); err != nil {
		t.Fatal(err)
	}
	got := s.List()
	if got[0].ID != b.ID || got[1].ID != a.ID {
		t.Fatalf("expected B before A, got %+v", got)
	}
	if err := s.MoveActiveAmong(ActiveIDsInOrder(s.List()), b.ID, 1); err != nil {
		t.Fatal(err)
	}
	got = s.List()
	if got[0].ID != a.ID || got[1].ID != b.ID {
		t.Fatalf("expected A before B again, got %+v", got)
	}
	if err := s.MoveActiveAmong(ActiveIDsInOrder(s.List()), c.ID, 1); err != nil {
		t.Fatal(err)
	}
}

func TestMoveActiveAmongWithTagFilter(t *testing.T) {
	s, _ := memStore()
	a := s.Add("A", "x")
	s.Add("B", "y")
	c := s.Add("C", "x")

	visible := FilterByTag(s.List(), "x")
	activeIDs := ActiveIDsInOrder(visible)
	if len(activeIDs) != 2 {
		t.Fatalf("expected 2 visible active tasks, got %v", activeIDs)
	}
	if err := s.MoveActiveAmong(activeIDs, c.ID, -1); err != nil {
		t.Fatal(err)
	}
	got := s.List()
	aIdx, cIdx := -1, -1
	for i, task := range got {
		switch task.ID {
		case a.ID:
			aIdx = i
		case c.ID:
			cIdx = i
		}
	}
	if cIdx < 0 || aIdx < 0 || cIdx >= aIdx {
		t.Fatalf("expected C before A in global list, got %+v", got)
	}
}

func TestMoveActiveAmongSkipsCompleted(t *testing.T) {
	s, _ := memStore()
	a := s.Add("A", "")
	b := s.Add("B", "")
	if err := s.SetCompleted(a.ID, true); err != nil {
		t.Fatal(err)
	}
	err := s.MoveActiveAmong(ActiveIDsInOrder(s.List()), a.ID, -1)
	if err == nil {
		t.Fatal("expected error moving completed task")
	}
	if len(ActiveIDsInOrder(s.List())) != 1 || ActiveIDsInOrder(s.List())[0] != b.ID {
		t.Fatalf("only B should remain active: %+v", s.List())
	}
}

func TestAddPlacesNewTaskInActiveSection(t *testing.T) {
	s, _ := memStore()
	done := s.Add("done", "")
	if err := s.SetCompleted(done.ID, true); err != nil {
		t.Fatal(err)
	}
	active := s.Add("new", "")
	all := s.List()
	if len(all) != 2 || all[0].ID != active.ID || all[1].ID != done.ID {
		t.Fatalf("new task should be active before completed: %+v", all)
	}
}

func TestTaskByID(t *testing.T) {
	all := []Task{{ID: "1", Name: "a"}}
	got, ok := TaskByID(all, "1")
	if !ok || got.Name != "a" {
		t.Fatalf("unexpected task: %+v %v", got, ok)
	}
	if _, ok := TaskByID(all, "missing"); ok {
		t.Fatal("expected missing task")
	}
}

func TestSetCompletedIdempotent(t *testing.T) {
	s, raw := memStore()
	t1 := s.Add("One", "t")
	if err := s.SetCompleted(t1.ID, true); err != nil {
		t.Fatal(err)
	}
	before := *raw
	if err := s.SetCompleted(t1.ID, true); err != nil {
		t.Fatal(err)
	}
	if before != *raw {
		t.Fatal("setting completed twice should keep stable order")
	}
}
