package mr

import (
	"fmt"
	"time"

	"branchy/internal/gitlab"
	"branchy/internal/project"
	"branchy/internal/tree"
)

// Action is the outcome of one merge-request attempt.
type Action string

const (
	ActionCreated Action = "created"
	ActionSkipped Action = "skipped"
	ActionFailed  Action = "failed"
)

// SkipReason distinguishes skipped subtypes. It is contributor-facing, not printed.
type SkipReason string

const (
	SkipNone         SkipReason = ""
	SkipAlreadyOpen  SkipReason = "already_open"
	SkipUserDeclined SkipReason = "user_declined"
)

// CreateRequest holds parameters for a single MR creation attempt.
type CreateRequest struct {
	Source      string
	Target      string
	Title       string
	Description string
}

// CreateResult summarizes one MR attempt.
type CreateResult struct {
	Source     string
	Target     string
	Action     Action
	SkipReason SkipReason
	URL        string
	Message    string
}

// DefaultTitle returns the standard MR title for a source→target pair.
func DefaultTitle(source, target string) string {
	return fmt.Sprintf("MR: %s → %s", source, target)
}

// DefaultDescription returns a timestamped MR description.
func DefaultDescription() string {
	ts := time.Now().Format("2006-01-02 15:04:05 -0700")
	return fmt.Sprintf("Manual merge request created by branchy on %s.", ts)
}

// ValidateBranches checks tree membership and that source and target differ.
func ValidateBranches(doc *tree.Document, source, target string) error {
	if source == "" || target == "" {
		return fmt.Errorf("source and target are required")
	}
	if source == target {
		return fmt.Errorf("source and target must differ")
	}
	if _, ok := doc.Branches[source]; !ok {
		return fmt.Errorf("branch %q not in tree", source)
	}
	if _, ok := doc.Branches[target]; !ok {
		return fmt.Errorf("branch %q not in tree", target)
	}
	return nil
}

// Create validates branches, checks for an existing open MR, and creates one if needed.
// It does not open a browser; callers handle that.
//
// Return matrix:
//   - validation (empty, equal, or unknown branch) → (nil, err); GitLab is not called
//   - auth failure → (nil, err)
//   - already-open (found or recovered after create race) → skipped + SkipAlreadyOpen, nil error
//   - unrecovered GitLab create failure → failed result, nil error
//   - created → created result, nil error
func Create(p *project.Project, req CreateRequest) (*CreateResult, error) {
	if err := ValidateBranches(p.Tree, req.Source, req.Target); err != nil {
		return nil, err
	}

	client := &gitlab.Client{Dir: p.Path}
	if err := client.AuthOK(); err != nil {
		return nil, fmt.Errorf("glab auth: %w (run: glab auth login)", err)
	}

	title := req.Title
	if title == "" {
		title = DefaultTitle(req.Source, req.Target)
	}
	desc := req.Description
	if desc == "" {
		desc = DefaultDescription()
	}

	res := &CreateResult{Source: req.Source, Target: req.Target}

	existing, err := client.FindOpenMR(req.Source, req.Target)
	if err != nil {
		res.Action = ActionFailed
		res.Message = err.Error()
		return res, nil
	}
	if existing != "" {
		res.Action = ActionSkipped
		res.SkipReason = SkipAlreadyOpen
		res.URL = existing
		res.Message = "open MR already exists"
		return res, nil
	}

	url, err := client.CreateMR(req.Source, req.Target, title, desc)
	if err != nil {
		// Race or glab quirk: treat recovered existing open MR as skip, not failure.
		if gitlab.IsAlreadyExists(err) && url != "" {
			res.Action = ActionSkipped
			res.SkipReason = SkipAlreadyOpen
			res.URL = url
			res.Message = "open MR already exists"
			return res, nil
		}
		if existing, findErr := client.FindOpenMR(req.Source, req.Target); findErr == nil && existing != "" {
			res.Action = ActionSkipped
			res.SkipReason = SkipAlreadyOpen
			res.URL = existing
			res.Message = "open MR already exists"
			return res, nil
		}
		res.Action = ActionFailed
		res.Message = err.Error()
		return res, nil
	}

	res.Action = ActionCreated
	res.URL = url
	return res, nil
}
