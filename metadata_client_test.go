/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kowabunga-cloud/common/ws"
)

func TestGetInstanceMetadataRawSendsAgentCredentials(t *testing.T) {
	var headers http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		_, _ = w.Write([]byte(`{"id": "i"}`))
	}))
	defer srv.Close()

	settings := CloudInitInstanceMetaDataSettings{
		InstanceID:       "instance-id",
		LocalIP:          "10.0.0.2",
		MetadataEndpoint: srv.URL,
	}
	if _, err := GetInstanceMetadataRaw(settings); err != nil {
		t.Fatal(err)
	}
	if headers.Get(HttpHeaderKowabungaInstanceID) != "instance-id" || headers.Get(HttpHeaderKowabungaSourceIP) != "10.0.0.2" {
		t.Errorf("missing instance headers: %v", headers)
	}
	if headers.Get(ws.WsHeaderAgentId) != "" || headers.Get(ws.WsHeaderAgentApiKey) != "" {
		t.Error("no agent credentials should be sent for regular instances")
	}

	settings.ControllerAgentID = "agent-id"
	settings.ControllerAgentToken = "secret"
	if _, err := GetInstanceMetadataRaw(settings); err != nil {
		t.Fatal(err)
	}
	if headers.Get(ws.WsHeaderAgentId) != "agent-id" || headers.Get(ws.WsHeaderAgentApiKey) != "secret" {
		t.Errorf("agent credentials not sent: %v", headers)
	}
}
