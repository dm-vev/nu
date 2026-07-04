package message

import "testing"

func TestAppendTextCoalescesAdjacentTextParts(t *testing.T) {
	msg := NewAssistant()
	msg.AppendText("hel")
	msg.AppendText("lo")

	if len(msg.Parts) != 1 || msg.Parts[0].Text != "hello" {
		t.Fatalf("parts = %#v, want one coalesced text part", msg.Parts)
	}
}

func TestThinkingAndToolPartsKeepOrdering(t *testing.T) {
	msg := NewAssistant()
	msg.AppendThinking("plan")
	msg.AddTool("call-1", "bash", `{"command":"pwd"}`)
	msg.FinishTool("call-1", "/tmp\n", false)
	msg.AppendText("done")

	if len(msg.Parts) != 3 {
		t.Fatalf("parts = %#v, want thinking/tool/text", msg.Parts)
	}
	if msg.Parts[0].Kind != PartThinking || msg.Parts[1].ToolState != ToolSuccess || msg.Parts[2].Text != "done" {
		t.Fatalf("parts = %#v, want ordered thinking/tool/text", msg.Parts)
	}
}

func TestReplaceTextDoesNotDeleteToolParts(t *testing.T) {
	msg := NewAssistant()
	msg.AddTool("call-1", "edit", "{}")
	msg.AppendText("partial")

	msg.ReplaceText("final")

	if len(msg.Parts) != 2 || msg.Parts[0].Kind != PartTool || msg.Parts[1].Text != "final" {
		t.Fatalf("parts = %#v, want tool preserved and text replaced", msg.Parts)
	}
}
