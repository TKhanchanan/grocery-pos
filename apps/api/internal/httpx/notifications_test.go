package httpx

import (
	"context"
	"errors"
	"testing"

	"grocery-pos/apps/api/internal/line"
)

type fakeLinePusher struct {
	messages []line.Message
	errors   []error
}

func (f *fakeLinePusher) Push(_ context.Context, _, _ string, messages ...line.Message) error {
	f.messages = append(f.messages, messages...)
	if len(f.errors) == 0 {
		return nil
	}
	err := f.errors[0]
	f.errors = f.errors[1:]
	return err
}

func TestPushLineMessageFallsBackToText(t *testing.T) {
	client := &fakeLinePusher{errors: []error{errors.New("flex rejected"), nil}}
	flex := line.BuildTestNotification(line.TestNotificationInput{})

	err := pushLineMessage(context.Background(), client, "token", "target", "LINE_TEST", "fallback text", flex)
	if err != nil {
		t.Fatal(err)
	}
	if len(client.messages) != 2 {
		t.Fatalf("push count = %d, want 2", len(client.messages))
	}
	if _, ok := client.messages[0].(line.FlexMessage); !ok {
		t.Errorf("first message type = %T, want line.FlexMessage", client.messages[0])
	}
	text, ok := client.messages[1].(line.TextMessage)
	if !ok {
		t.Fatalf("second message type = %T, want line.TextMessage", client.messages[1])
	}
	if text.Text != "fallback text" {
		t.Errorf("fallback text = %q", text.Text)
	}
}
