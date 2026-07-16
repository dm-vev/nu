package cli

import (
	"strings"
	"testing"
)

func TestCLIHelpMentionsCoreModes(t *testing.T) {
	help := Help(nil)
	for _, want := range []string{"--print", "--mode json", "--mode rpc", "package", "share"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help() missing %q in:\n%s", want, help)
		}
	}
}

func TestCLIHelpIncludesExtensionFlags(t *testing.T) {
	help := Help([]ExtensionFlag{{Name: "--x-test", Description: "test flag", Source: "test-ext"}})
	for _, want := range []string{"--x-test", "test flag", "test-ext"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help() missing extension text %q in:\n%s", want, help)
		}
	}
}
