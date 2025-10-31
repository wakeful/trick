// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package broadcast

import (
	"log/slog"
	"strings"
	"sync"
)

type Message struct {
	Chain string
	Role  string
}

func (m *Message) String() string {
	var builder strings.Builder
	builder.WriteString("event: jump\n")
	builder.WriteString("data: chain: ")
	builder.WriteString(m.Chain)
	builder.WriteString("\n")
	builder.WriteString("data: role: ")
	builder.WriteString(m.Role)
	builder.WriteString("\n\n")

	return builder.String()
}

type Broadcaster struct {
	mu      sync.RWMutex
	subs    map[chan Message]struct{}
	lastMsg *Message
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		mu:      sync.RWMutex{},
		subs:    make(map[chan Message]struct{}),
		lastMsg: nil,
	}
}

func (b *Broadcaster) Subscribe() (<-chan Message, func()) {
	const buffer = 10

	target := make(chan Message, buffer)

	b.mu.Lock()
	b.subs[target] = struct{}{}

	if b.lastMsg != nil {
		select {
		case target <- *b.lastMsg:
			slog.Debug("sent current role to new subscriber")
		default:
			slog.Debug("failed to send current role to new subscriber")
		}
	}

	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()

		if _, ok := b.subs[target]; ok {
			delete(b.subs, target)
			close(target)
		}

		b.mu.Unlock()
	}

	return target, unsub
}

func (b *Broadcaster) Publish(msg Message) {
	b.mu.Lock()
	b.lastMsg = &msg
	b.mu.Unlock()

	b.mu.RLock()

	for ch := range b.subs {
		select {
		case ch <- msg:
			slog.Debug("message delivered to subscriber")
		default:
			slog.Debug("message dropped")
		}
	}

	b.mu.RUnlock()
}
