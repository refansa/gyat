// Package pager provides an internal, cross-platform pager implementation
// used by the gyat CLI to present long-form output interactively on TTYs.
//
// The package is intentionally internal: consumers should use the public API
// exposed here (NewPager, Render, etc.) rather than depending on platform
// specific binaries. Initial skeleton created for Phase 1 setup tasks.
package pager
