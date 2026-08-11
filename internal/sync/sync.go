package sync

import (
	"fmt"
	"strings"
	"time"

	"branchy/internal/browser"
	"branchy/internal/gitlab"
	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/tree"
)

// Result summarizes one MR attempt.
type Result struct {
	Parent  string
	Child   string
	Action  string // created, skipped, failed
	URL     string
	Message string
}

// Summary aggregates a sync run.
type Summary struct {
	Results []Result
}

// Options configures an MR sync run.
type Options struct {
	FromBranch string
	Confirm    func(parent, child string) (bool, error)
}

var browserOpen = browser.Open

// Run creates MRs for each parent→child edge below FromBranch.
func Run(p *project.Project, opts Options) (*Summary, error) {
	if opts.FromBranch == "" {
		return nil, fmt.Errorf("from branch is required")
	}
	if _, ok := p.Tree.Branches[opts.FromBranch]; !ok {
		return nil, fmt.Errorf("branch %q not in tree", opts.FromBranch)
	}

	client := &gitlab.Client{Dir: p.Path}
	if err := client.AuthOK(); err != nil {
		return nil, fmt.Errorf("glab auth: %w (run: glab auth login)", err)
	}

	edges := p.Tree.CollectEdges(opts.FromBranch)
	if len(edges) == 0 {
		return &Summary{}, nil
	}

	summary := &Summary{}

	for _, edge := range edges {
		res := processEdge(p, edge, opts)
		summary.Results = append(summary.Results, res)
	}

	return summary, nil
}

// RunEdge creates an MR for a single edge without confirmation.
func RunEdge(p *project.Project, edge tree.Edge) Result {
	return processEdge(p, edge, Options{})
}

// CreatedURLs returns MR URLs for created results in DFS order.
func CreatedURLs(summary *Summary) []string {
	if summary == nil {
		return nil
	}
	var urls []string
	for _, r := range summary.Results {
		if r.Action == mr.ActionCreated && r.URL != "" {
			urls = append(urls, r.URL)
		}
	}
	return urls
}

// OpenableURLs returns MR URLs for confirmed edges in DFS order: newly created
// MRs and edges skipped because an open MR already existed. User-declined and
// failed edges without a URL are excluded.
func OpenableURLs(summary *Summary) []string {
	if summary == nil {
		return nil
	}
	var urls []string
	for _, r := range summary.Results {
		if r.URL == "" {
			continue
		}
		switch r.Action {
		case mr.ActionCreated:
			urls = append(urls, r.URL)
		case mr.ActionSkipped:
			if r.Message == "skipped by user" {
				continue
			}
			urls = append(urls, r.URL)
		}
	}
	return urls
}

// OpenURLs opens each URL sequentially in order. Returns a non-fatal warning if any open fails.
func OpenURLs(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	var warnings []string
	for i, url := range urls {
		if i > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		if err := browserOpen(url); err != nil {
			warnings = append(warnings, fmt.Sprintf("could not open %s: %v", url, err))
		}
	}
	return strings.Join(warnings, "; ")
}

func processEdge(p *project.Project, edge tree.Edge, opts Options) Result {
	res := Result{Parent: edge.Parent, Child: edge.Child}

	if opts.Confirm != nil {
		ok, err := opts.Confirm(edge.Parent, edge.Child)
		if err != nil {
			res.Action = "failed"
			res.Message = err.Error()
			return res
		}
		if !ok {
			res.Action = "skipped"
			res.Message = "skipped by user"
			return res
		}
	}

	ts := time.Now().Format("2006-01-02 15:04:05 -0700")
	mrRes, err := mr.Create(p, mr.CreateRequest{
		Source:      edge.Parent,
		Target:      edge.Child,
		Title:       fmt.Sprintf("Sync: %s → %s", edge.Parent, edge.Child),
		Description: fmt.Sprintf("Automated branch sync created by branchy on %s.", ts),
	})
	if err != nil {
		res.Action = "failed"
		res.Message = err.Error()
		return res
	}

	res.Action = mrRes.Action
	res.URL = mrRes.URL
	res.Message = mrRes.Message
	return res
}
