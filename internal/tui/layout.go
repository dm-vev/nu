package tui

import "github.com/dm-vev/nu/internal/tui/components"

func (a *App) buildLayout() {
	a.ui.AddChild(a.header)
	a.ui.AddChild(a.chat)
	a.ui.AddChild(components.NewFill())
	a.ui.AddChild(a.models)
	a.ui.AddChild(a.menu)
	a.ui.AddChild(a.status)
	a.ui.AddChild(a.editor)
	a.ui.AddChild(a.footer)
}

func commandMenuOptions() components.CommandMenuOptions {
	return components.CommandMenuOptions{
		MaxItems: 8,
		Text:     ansiText,
		Accent:   greenBold,
		Muted:    muted,
	}
}

func modelMenuOptions() components.ModelMenuOptions {
	return components.ModelMenuOptions{
		MaxVisible: 10,
		Text:       ansiText,
		Accent:     greenBold,
		Muted:      muted,
		Success:    greenBold,
		Border:     green,
		Error:      red,
	}
}

func headerOptions(opts AppOptions) components.HeaderOptions {
	separator := " · "
	if limitedCharset(opts) {
		separator = " | "
	}
	return components.HeaderOptions{
		AppName:       "Nu",
		Version:       opts.Version,
		Accent:        greenBold,
		Dim:           dim,
		Muted:         muted,
		HelpSeparator: separator,
		PaddingX:      1,
		PaddingY:      1,
	}
}

func footerOptions(opts AppOptions) components.FooterOptions {
	return components.FooterOptions{
		CWD:         opts.CWD,
		Home:        opts.Home,
		Branch:      firstNonEmpty(opts.Branch, currentGitBranch(opts.CWD)),
		Provider:    opts.Provider,
		Model:       firstNonEmpty(opts.ModelLabel, opts.Model),
		Context:     opts.Context,
		Dim:         dim,
		NoticeStyle: red,
	}
}
