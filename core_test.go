package main

import (
	"strings"
	"testing"
)

func TestSendRequest_EmptyURL(t *testing.T) {
	m := NewModel()
	m.urlInput.SetValue("") // empty URL

	cmd := sendRequest(m)
	resp := cmd().(responseMsg)

	if resp.err == nil || !strings.Contains(resp.err.Error(), "URL cannot be empty") {
		t.Errorf("Expected URL validation error, got: %v", resp.err)
	}
}

func TestSendRequest_GET_WithHeaders(t *testing.T) {
	m := NewModel()
	m.methodIndex = 0 // GET
	m.urlInput.SetValue("https://httpbin.org/get")

	// Add a header
	m.headers[0].key.SetValue("Test-Key")
	m.headers[0].value.SetValue("123")

	cmd := sendRequest(m)
	resp := cmd().(responseMsg)

	if resp.err != nil {
		t.Errorf("Expected no error, got: %v", resp.err)
	}
	if !strings.Contains(resp.body, "Test-Key") {
		t.Errorf("Expected response to include header key, got: %s", resp.body)
	}
}

func TestSendRequest_POST_WithBody(t *testing.T) {
	m := NewModel()
	m.methodIndex = 1 // POST
	m.urlInput.SetValue("https://httpbin.org/post")
	m.bodyInput.SetValue(`{"foo":"bar"}`)

	cmd := sendRequest(m)
	resp := cmd().(responseMsg)

	if resp.err != nil {
		t.Errorf("Expected no error, got: %v", resp.err)
	}
	if !strings.Contains(resp.body, `"foo": "bar"`) {
		t.Errorf("Expected response to include body content, got: %s", resp.body)
	}
}
