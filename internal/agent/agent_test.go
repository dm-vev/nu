package agent

import (
	"context"
	"errors"
	"testing"

	"nu/internal/provider"
	"nu/internal/testkit"
)

func TestNUF050TextOnlyTurnEnds(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "hello"},
		provider.Event{Type: provider.EventDone, StopReason: "stop"},
	)
	var events []Event
	a := New(Options{
		Provider: fake,
		Emit: func(ev Event) {
			events = append(events, ev)
		},
	})

	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}
	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("Provider requests = %d, want 1", len(requests))
	}
	if requests[0].Messages[0].Content != "hi" {
		t.Fatalf("Provider prompt = %q, want hi", requests[0].Messages[0].Content)
	}
	if got := events[len(events)-1].Type; got != "turn_end" {
		t.Fatalf("Last event = %q, want turn_end", got)
	}
}

func TestNUF050AbortStopsProviderAndTools(t *testing.T) {
	fake := &blockingProvider{started: make(chan struct{})}
	a := New(Options{Provider: fake})

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Prompt(context.Background(), Prompt{Text: "hi"})
	}()
	<-fake.started

	a.Abort()
	err := <-errCh
	if !errors.Is(err, provider.ErrStream) {
		t.Fatalf("Prompt error = %v, want provider ErrStream", err)
	}
}

type blockingProvider struct {
	started chan struct{}
}

func (p *blockingProvider) Stream(ctx context.Context, _ provider.Request) (<-chan provider.Event, error) {
	close(p.started)
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}
