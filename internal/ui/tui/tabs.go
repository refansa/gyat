package tui

import (
	"github.com/refansa/gyat/v2/internal/ui/components"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

var normalizeEntry = components.NormalizeEntry
var renderDetail = components.RenderDetail
var renderInputPanel = components.RenderInputPanel

func moveActiveTab(entry *uiModel.RepositoryEntry, delta int) {
	components.MoveActiveTab(entry, delta)
}
