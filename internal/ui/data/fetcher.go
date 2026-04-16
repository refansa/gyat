package data

import (
	"context"
	"sync"

	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

// SummaryFetchFunc loads one repository entry for the given workspace path.
type SummaryFetchFunc func(context.Context, string) (uiModel.RepositoryEntry, error)

// Result is one asynchronous repository fetch result.
type Result struct {
	Path  string
	Entry uiModel.RepositoryEntry
	Err   error
}

// FetchRepositoryEntriesAsync fetches repository entries concurrently and emits
// results as they become available.
func FetchRepositoryEntriesAsync(ctx context.Context, paths []string, fetch SummaryFetchFunc) <-chan Result {
	results := make(chan Result)
	if len(paths) == 0 || fetch == nil {
		close(results)
		return results
	}

	var wg sync.WaitGroup
	wg.Add(len(paths))
	for _, path := range paths {
		path := path
		go func() {
			defer wg.Done()
			entry, err := fetch(ctx, path)
			select {
			case <-ctx.Done():
				return
			case results <- Result{Path: path, Entry: entry, Err: err}:
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// CollectRepositoryEntries preserves the input order while fetching entries in
// parallel so callers can do the work outside the Bubble Tea update loop.
func CollectRepositoryEntries(ctx context.Context, paths []string, fetch SummaryFetchFunc) ([]uiModel.RepositoryEntry, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	ordered := make([]uiModel.RepositoryEntry, len(paths))
	indexByPath := make(map[string][]int, len(paths))
	for index, path := range paths {
		indexByPath[path] = append(indexByPath[path], index)
	}

	var firstErr error
	for result := range FetchRepositoryEntriesAsync(ctx, paths, fetch) {
		indices := indexByPath[result.Path]
		if len(indices) == 0 {
			continue
		}
		index := indices[0]
		indexByPath[result.Path] = indices[1:]
		ordered[index] = result.Entry
		if firstErr == nil && result.Err != nil {
			firstErr = result.Err
		}
	}

	return ordered, firstErr
}
