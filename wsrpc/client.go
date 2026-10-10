/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package wsrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kowabunga-cloud/common/klog"

	"github.com/gorilla/websocket"
)

type WsRpcClient struct {
	mutex sync.Mutex // protects following
	seq   uint64
	conn  *websocket.Conn

	response chan []byte

	// Terminated receives a single value once the connection is over.
	//
	// Deprecated: use Done(), which can be waited on by any number of
	// goroutines.
	Terminated chan bool

	done          chan struct{}
	terminateOnce sync.Once
	debug         bool
}

func NewWsRpcClient(conn *websocket.Conn, debug bool) *WsRpcClient {
	client := WsRpcClient{
		conn:       conn,
		debug:      debug,
		response:   make(chan []byte, 1),
		Terminated: make(chan bool, 1),
		done:       make(chan struct{}),
	}

	// register pong handler, keeping websocket alive
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	// start a background infinite go-routine, handling all messages
	// stop it when underlying connection's broken
	go client.pumpMessages()

	// start a background infinite go-routine, pinging remote peer for keepalive
	go client.keepalive()

	return &client
}

// Done returns a channel that is closed once the connection is over, whether
// it's been closed locally or remotely, or broken.
func (client *WsRpcClient) Done() <-chan struct{} {
	return client.done
}

// terminate closes the underlying connection and signals termination, only
// once, however many routines detect it.
func (client *WsRpcClient) terminate() {
	client.terminateOnce.Do(func() {
		close(client.done)
		err := client.conn.Close()
		if err != nil {
			klog.Debugf("wsrpc: closing connection: %v", err)
		}
		// buffered, sent only once: never blocks
		client.Terminated <- true
	})
}

func (client *WsRpcClient) pumpMessages() {
	defer client.terminate()

	for {
		// ensure we have a proper connection
		if client.conn == nil {
			klog.Errorf("wsrpc: no underlying WebSocket connection")
			return
		}

		// pump incoming messages
		mt, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				klog.Infof("wsrpc: unexpected remote client connection closure detected")
			}
			return
		}

		switch mt {
		case websocket.BinaryMessage:
			// deliver pumped message to channel, ready to be processed by caller
			select {
			case client.response <- message:
			case <-client.done:
				return
			}
		default:
			klog.Infof("Handling other type of message: %d", mt)
		}
	}
}

func (client *WsRpcClient) keepalive() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-client.done:
			return
		case <-ticker.C:
			// WriteControl is safe for use concurrently with RPC requests writes
			err := client.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			if err != nil {
				klog.Errorf("remote WebSocket disconnection detected")
				client.terminate()
				return
			}
		}
	}
}

// Close closes the connection, terminating any pending call.
func (client *WsRpcClient) Close() error {
	client.terminate()
	return nil
}

func (client *WsRpcClient) newRequest(serviceMethod string, args any) *RpcRequest {
	seq := client.seq
	client.seq++

	req := RpcRequest{
		Id:            seq,
		ServiceMethod: serviceMethod,
		Params:        args,
	}

	if client.debug {
		klog.Debugf("WsRpcRequest: %#v", req)
	}

	return &req
}

func (client *WsRpcClient) sendRequest(req *RpcRequest) error {
	msg, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// write RPC request to server
	err = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		client.terminate()
		return err
	}
	err = client.conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		client.terminate()
		return err
	}

	return nil
}

func (client *WsRpcClient) getResponse(id uint64, reply any) (*RpcResponse, error) {
	timeout := time.After(WsCloseResponseTimeout)

	// dequeue (with timeout) messages from handler, whichever comes first
	for {
		select {
		case message := <-client.response:
			// discard late responses to previous (timed out) requests
			var header struct {
				Id uint64 `json:"id"`
			}
			err := json.Unmarshal(message, &header)
			if err != nil {
				return nil, err
			}
			if header.Id != id {
				klog.Infof("wsrpc: discarding stale response to request %d", header.Id)
				continue
			}

			resp := RpcResponse{
				Result: reply, // pointer to reply, for unmarshalling into
			}

			// read RPC reply from server
			err = json.Unmarshal(message, &resp)
			if err != nil {
				return &resp, err
			}

			if client.debug {
				klog.Debugf("WsRpcResponse: %#v", resp)
			}

			return &resp, nil
		case <-client.done:
			return nil, errors.New("wsrpc: connection closed")
		case <-timeout:
			return nil, errors.New("wsrpc: timeout")
		}
	}
}

func (client *WsRpcClient) Call(serviceMethod string, args any, reply any) error {

	// prevent concurential calls from happening on the same WebSocket transport layer
	client.mutex.Lock()
	defer client.mutex.Unlock()

	req := client.newRequest(serviceMethod, args)
	err := client.sendRequest(req)
	if err != nil {
		klog.Errorf("unable to send RPC request: %s", err)
		return err
	}

	resp, err := client.getResponse(req.Id, reply)
	if err != nil {
		klog.Errorf("unable to read RPC response: %s", err)
		return err
	}

	if resp.Id != req.Id {
		err := errors.New("wsrpc: mismatch between RPC request and response IDs")
		klog.Error(err)
		return err
	}

	if resp.Error != nil {
		err := fmt.Errorf("wsrpc: %s() error (%v)", serviceMethod, resp.Error)
		klog.Error(err)
		return err
	}

	return nil
}
