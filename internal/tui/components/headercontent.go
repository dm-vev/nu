package components

import "strings"

func (h *Header) content() string {
	logo := h.opts.Accent(h.opts.AppName)
	if strings.TrimSpace(h.opts.Version) != "" {
		logo += h.opts.Dim(" v" + strings.TrimSpace(h.opts.Version))
	}
	if h.expanded {
		return strings.Join([]string{
			logo,
			h.expandedHelp(),
			"",
			h.onboarding(),
			"",
			h.contextBlock(),
		}, "\n")
	}
	return strings.Join([]string{
		logo,
		h.compactHelp(),
		h.opts.Dim("Press ctrl+o to show full startup help and loaded resources."),
		"",
		h.onboarding(),
		"",
		h.contextBlock(),
	}, "\n")
}

func (h *Header) compactHelp() string {
	parts := []string{
		h.opts.Dim("esc") + h.opts.Muted(" interrupt"),
		h.opts.Dim("ctrl+c/d") + h.opts.Muted(" clear/exit"),
		h.opts.Dim("/") + h.opts.Muted(" commands"),
		h.opts.Dim("!") + h.opts.Muted(" bash"),
		h.opts.Dim("ctrl+o") + h.opts.Muted(" more"),
	}
	return strings.Join(parts, h.opts.Muted(h.opts.HelpSeparator))
}

func (h *Header) expandedHelp() string {
	parts := []string{
		h.opts.Dim("escape") + h.opts.Muted(" to interrupt"),
		h.opts.Dim("ctrl+c") + h.opts.Muted(" to clear"),
		h.opts.Dim("ctrl+c twice") + h.opts.Muted(" to exit"),
		h.opts.Dim("ctrl+d") + h.opts.Muted(" to exit when empty"),
		h.opts.Dim("/") + h.opts.Muted(" for commands"),
		h.opts.Dim("!") + h.opts.Muted(" to run bash"),
		h.opts.Dim("ctrl+o") + h.opts.Muted(" to expand tools/help"),
	}
	return strings.Join(parts, "\n")
}

func (h *Header) onboarding() string {
	return h.opts.Dim("Nu can explain its own features and look up its docs. Ask it how to use or extend Nu.")
}

func (h *Header) contextBlock() string {
	return h.opts.Accent("[Context]") + "\n" + h.opts.Dim("  AGENTS.md")
}
