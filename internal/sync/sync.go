package sync

import (
	"fmt"
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
	OnStatus   func(msg string)
}

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
	var urls []string

	for _, edge := range edges {
		res := processEdge(p, edge, opts)
		summary.Results = append(summary.Results, res)
		if res.URL != "" && res.Action != "failed" {
			urls = append(urls, res.URL)
		}
	}

	if len(urls) > 0 && opts.OnStatus != nil {
		opts.OnStatus(fmt.Sprintf("Opening %d MR(s) in browser...", len(urls)))
	}
	for _, url := range urls {
		_ = browser.Open(url)
	}

	return summary, nil
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
