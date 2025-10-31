// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package broadcast_test

import (
	"testing"
	"time"

	"github.com/wakeful/trick/internal/broadcast"
)

func TestMessage_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  broadcast.Message
		expected string
	}{
		{
			name: "basic message",
			message: broadcast.Message{
				Chain: "test-chain",
				Role:  "test-role",
			},
			expected: "event: jump\ndata: chain: test-chain\ndata: role: test-role\n\n",
		},
		{
			name: "empty message",
			message: broadcast.Message{
				Chain: "",
				Role:  "",
			},
			expected: "event: jump\ndata: chain: \ndata: role: \n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.message.String()
			if result != tt.expected {
				t.Errorf("Message.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestBroadcaster_Publish(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()
	msg := broadcast.Message{Chain: "test", Role: "role"}
	b.Publish(msg)

	ch1, unsub1 := b.Subscribe()
	defer unsub1()

	ch2, unsub2 := b.Subscribe()
	defer unsub2()

	b.Publish(msg)

	select {
	case received := <-ch1:
		if received.Chain != msg.Chain || received.Role != msg.Role {
			t.Errorf("ch1 received %+v, expected %+v", received, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 did not receive message within timeout")
	}

	select {
	case received := <-ch2:
		if received.Chain != msg.Chain || received.Role != msg.Role {
			t.Errorf("ch2 received %+v, expected %+v", received, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 did not receive message within timeout")
	}
}

func TestBroadcaster_PublishWithFullBuffer(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	ch, unsub := b.Subscribe()
	defer unsub()

	const numMessages = 10
	for range numMessages {
		msg := broadcast.Message{Chain: "test", Role: "role"}
		b.Publish(msg)
	}

	extraMsg := broadcast.Message{Chain: "extra", Role: "extra"}
	b.Publish(extraMsg)

	receivedCount := 0

	for range numMessages {
		select {
		case <-ch:
			receivedCount++
		case <-time.After(100 * time.Millisecond):
			return
		}
	}

	if receivedCount != numMessages {
		t.Errorf("Expected to receive 10 messages, got %d", receivedCount)
	}

	select {
	case msg := <-ch:
		t.Errorf("Received unexpected message: %+v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}
