package tui

import (
	"fmt"
	"strings"

	"github.com/dm-vev/nu/internal/model"
)

func (a *App) scopedModelsCommandText() string {
	a.mu.Lock()
	models := append([]modelSummary(nil), summarizeModels(a.available, a.provider, a.modelID)...)
	a.mu.Unlock()
	if len(models) == 0 {
		return "Scoped models\n\nNo models are visible for the current provider credentials."
	}
	var b strings.Builder
	b.WriteString("Scoped models\n\n| Current | Provider | Model | Display |\n| --- | --- | --- | --- |\n")
	for _, item := range models {
		current := ""
		if item.Current {
			current = "*"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", current, item.Provider, item.ID, item.Display)
	}
	return b.String()
}

type modelSummary struct {
	Provider string
	ID       string
	Display  string
	Current  bool
}

func summarizeModels(models []model.Model, providerID string, modelID string) []modelSummary {
	out := make([]modelSummary, 0, len(models))
	for _, entry := range models {
		out = append(out, modelSummary{
			Provider: entry.Provider,
			ID:       entry.ID,
			Display:  modelDisplayName(entry),
			Current:  entry.Provider == providerID && entry.ID == modelID,
		})
	}
	return out
}
