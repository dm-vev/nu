package slash

import (
	"strings"
	"testing"
)

func TestBuiltinsCopiesPiCommandSet(t *testing.T) {
	var names []string
	for _, command := range Builtins() {
		names = append(names, command.Name)
	}
	want := "settings,model,scoped-models,export,import,share,copy,name,session,changelog,hotkeys,fork,clone,tree,trust,login,logout,new,compact,resume,reload,quit"
	if strings.Join(names, ",") != want {
		t.Fatalf("commands = %s, want %s", strings.Join(names, ","), want)
	}
}

func TestParseSlashCommand(t *testing.T) {
	name, args, ok := Parse("/model fireworks/glm")
	if !ok || name != "model" || args != "fireworks/glm" {
		t.Fatalf("Parse = %q, %q, %v", name, args, ok)
	}
}

func TestFilterSlashCommands(t *testing.T) {
	matches := Filter("/mo", 3)
	if len(matches) != 2 || matches[0].Name != "model" || matches[1].Name != "scoped-models" {
		t.Fatalf("matches = %#v, want model and scoped-models", matches)
	}
}
