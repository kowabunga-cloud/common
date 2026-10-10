/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package wsrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type Echo struct{}

func (e *Echo) Say(args string, reply *string) error {
	*reply = "echo: " + args
	return nil
}

func dial(t *testing.T, handler http.HandlerFunc) *WsRpcClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	client := NewWsRpcClient(conn, false)
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

func echoServer(t *testing.T) http.HandlerFunc {
	upgrader := websocket.Upgrader{}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		server := NewWsRpcServer(conn, false)
		if err := server.Register(&Echo{}); err != nil {
			t.Error(err)
			return
		}
		_ = server.Listen()
	}
}

func TestConcurrentCalls(t *testing.T) {
	client := dial(t, echoServer(t))

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply string
			msg := fmt.Sprintf("hello %d", i)
			if err := client.Call("Echo.Say", msg, &reply); err != nil {
				t.Error(err)
				return
			}
			if reply != "echo: "+msg {
				t.Errorf("got reply %q for %q", reply, msg)
			}
		}(i)
	}
	wg.Wait()
}

func TestCloseTerminatesOnce(t *testing.T) {
	client := dial(t, echoServer(t))

	// closing several times, while the connection breaks, must not block
	done := make(chan struct{})
	go func() {
		_ = client.Close()
		_ = client.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Close() blocked")
	}

	select {
	case <-client.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Done() not closed after Close()")
	}

	// a single termination signal is delivered
	select {
	case <-client.Terminated:
	case <-time.After(5 * time.Second):
		t.Fatal("no termination signal")
	}
	select {
	case <-client.Terminated:
		t.Fatal("termination signalled twice")
	case <-time.After(100 * time.Millisecond):
	}

	// calls fail fast once closed
	var reply string
	if err := client.Call("Echo.Say", "late", &reply); err == nil {
		t.Error("call on a closed connection should fail")
	}
}

func TestStaleResponsesAreDiscarded(t *testing.T) {
	upgrader := websocket.Upgrader{}
	client := dial(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			_ = conn.Close()
		}()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req RpcRequest
			if err := json.Unmarshal(msg, &req); err != nil {
				return
			}
			// a late response to some previous request comes first
			stale, _ := json.Marshal(RpcResponse{Id: req.Id + 1000, Result: "stale"})
			resp, _ := json.Marshal(RpcResponse{Id: req.Id, Result: "fresh"})
			_ = conn.WriteMessage(websocket.BinaryMessage, stale)
			_ = conn.WriteMessage(websocket.BinaryMessage, resp)
		}
	})

	for range 3 {
		var reply string
		if err := client.Call("Any.Method", nil, &reply); err != nil {
			t.Fatal(err)
		}
		if reply != "fresh" {
			t.Fatalf("got %q, want the matching response", reply)
		}
	}
}
