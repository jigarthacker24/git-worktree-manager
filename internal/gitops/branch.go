package gitops

import (
	"fmt"
	"sort"
	"strings"
)

// BranchOption is a branch that can be checked out when adding a worktree.
type BranchOption struct {
	Display    string // label in the UI
	Ref        string // ref passed to git (local name or origin/branch)
	RemoteOnly bool   // true when Ref is a remote-tracking branch without a local branch
	LocalName  string // local branch name to create when RemoteOnly
}

// ListCheckoutBranches returns local branches and remote-only tracking branches.
func ListCheckoutBranches(repoPath string) ([]BranchOption, error) {
	localOut, err := runGit(repoPath, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	localSet := make(map[string]struct{})
	var opts []BranchOption
	for _, line := range strings.Split(strings.TrimSpace(localOut), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		localSet[name] = struct{}{}
		opts = append(opts, BranchOption{
			Display:   name,
			Ref:       name,
			LocalName: name,
		})
	}

	remoteOut, err := runGit(repoPath, "for-each-ref", "--format=%(refname:short)", "refs/remotes")
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(strings.TrimSpace(remoteOut), "\n") {
		ref := strings.TrimSpace(line)
		if ref == "" || strings.HasSuffix(ref, "/HEAD") {
			continue
		}
		slash := strings.Index(ref, "/")
		if slash <= 0 || slash >= len(ref)-1 {
			continue
		}
		localName := ref[slash+1:]
		if _, ok := localSet[localName]; ok {
			continue
		}
		opts = append(opts, BranchOption{
			Display:    ref,
			Ref:        ref,
			RemoteOnly: true,
			LocalName:  localName,
		})
	}

	sort.Slice(opts, func(i, j int) bool {
		return opts[i].Display < opts[j].Display
	})
	return opts, nil
}

// BranchOptionByDisplay finds an option by its UI label.
func BranchOptionByDisplay(opts []BranchOption, display string) (BranchOption, error) {
	for _, o := range opts {
		if o.Display == display {
			return o, nil
		}
	}
	return BranchOption{}, fmt.Errorf("branch not found: %s", display)
}

// DisplayNames returns sorted UI labels for branch selectors.
func DisplayNames(opts []BranchOption) []string {
	names := make([]string, len(opts))
	for i, o := range opts {
		names[i] = o.Display
	}
	return names
}

// PreferredBranchDisplay picks a default label from options.
func PreferredBranchDisplay(opts []BranchOption, preferred string) string {
	if preferred == "" {
		if len(opts) > 0 {
			return opts[0].Display
		}
		return ""
	}
	for _, o := range opts {
		if o.Display == preferred || o.Ref == preferred || o.LocalName == preferred {
			return o.Display
		}
	}
	remote := "origin/" + preferred
	for _, o := range opts {
		if o.Ref == remote {
			return o.Display
		}
	}
	if len(opts) > 0 {
		return opts[0].Display
	}
	return ""
}

// WorktreeAddArgs returns git arguments to add a worktree for an existing branch option.
func (o BranchOption) WorktreeAddArgs(wtPath string) []string {
	if o.RemoteOnly {
		return []string{"worktree", "add", "-b", o.LocalName, wtPath, o.Ref}
	}
	return []string{"worktree", "add", wtPath, o.Ref}
}

// SourceRefForNewBranch returns the start point for git worktree add -b.
func (o BranchOption) SourceRefForNewBranch() string {
	if o.Ref != "" {
		return o.Ref
	}
	return o.LocalName
}
