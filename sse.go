// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/wakeful/trick/internal/broadcast"
	"github.com/wakeful/trick/internal/ui"
)

func startSSEServer(
	ctx context.Context,
	broadcast *broadcast.Broadcaster,
	preRenderedHTML string,
) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler(preRenderedHTML))
	mux.HandleFunc("/events", events(broadcast))
	mux.HandleFunc("/static/mermaid.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		_, _ = w.Write(ui.MermaidScript)
	})

	const defaultAddr = "127.0.0.1:8742"

	serv := &http.Server{ //nolint:exhaustruct,gosec
		Addr:    defaultAddr,
		Handler: mux,
	}

	serverErr := make(chan error, 1)

	go func() {
		slog.Info("starting SSE server", slog.String("addr", defaultAddr))

		serverErr <- serv.ListenAndServe()

		slog.Info("SSE server stopped")
	}()

	for {
		select {
		case err := <-serverErr:
			slog.Error("SSE server error", slog.String("error", err.Error()))

			return

		case <-ctx.Done():
			slog.Debug("initiating SSE server shutdown")

			_ = serv.Shutdown(ctx)

			return
		}
	}
}

func events(
	broadcast *broadcast.Broadcaster,
) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Connection", "keep-alive")

		if f, ok := writer.(http.Flusher); ok {
			f.Flush()
		}

		target, unsubscribe := broadcast.Subscribe()
		defer unsubscribe()

		slog.Debug("SSE client connected", slog.String("client", request.RemoteAddr))

		notify := request.Context().Done()

		for {
			select {
			case message, ok := <-target:
				if !ok {
					return
				}

				_, _ = fmt.Fprint(writer, message.String())
				if f, ok := writer.(http.Flusher); ok {
					f.Flush()
				}

				slog.Debug(
					"SSE event delivered",
					slog.String("client", request.RemoteAddr),
					slog.String("chain", message.Chain),
					slog.String("role", message.Role),
				)
			case <-notify:
				slog.Debug("SSE client disconnected", slog.String("client", request.RemoteAddr))

				return
			}
		}
	}
}

func mainHandler(
	preRenderedHTML string,
) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(writer, preRenderedHTML)
	}
}
