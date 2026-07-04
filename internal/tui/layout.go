package tui

import (
	"nu/internal/tui/components/commandmenu"
	"nu/internal/tui/components/fill"
	"nu/internal/tui/components/footer"
	"nu/internal/tui/components/header"
	"nu/internal/tui/components/modelmenu"
)

func (a *App) buildLayout() {
	a.ui.AddChild(a.header)
	a.ui.AddChild(a.chat)
	a.ui.AddChild(fill.New())
	a.ui.AddChild(a.models)
	a.ui.AddChild(a.menu)
	a.ui.AddChild(a.status)
	a.ui.AddChild(a.editor)
	a.ui.AddChild(a.footer)
}

func commandMenuOptions() commandmenu.Options {
	return commandmenu.Options{
		MaxItems: 8,
		Text:     ansiText,
		Accent:   greenBold,
		Muted:    muted,
	}
}

func modelMenuOptions() modelmenu.Options {
	return modelmenu.Options{
		MaxVisible: 10,
		Text:       ansiText,
		Accent:     greenBold,
		Muted:      muted,
		Success:    greenBold,
		Border:     green,
		Error:      red,
	}
}

func headerOptions(opts AppOptions) header.Options {
	separator := " · "
	if limitedCharset(opts) {
		separator = " | "
	}
	return header.Options{
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

func footerOptions(opts AppOptions) footer.Options {
	return footer.Options{
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
