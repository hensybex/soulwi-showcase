package service

import "testing"

func TestBuildAPNSAlertConfig(t *testing.T) {
	cfg := buildAPNSAlertConfig()

	if cfg == nil {
		t.Fatalf("expected config, got nil")
	}

	if cfg.Headers == nil {
		t.Fatalf("expected headers to be set")
	}

	if got := cfg.Headers["apns-push-type"]; got != "alert" {
		t.Errorf("apns-push-type header = %q, want alert", got)
	}

	if got := cfg.Headers["apns-priority"]; got != "10" {
		t.Errorf("apns-priority header = %q, want 10", got)
	}

	if cfg.Payload == nil || cfg.Payload.Aps == nil {
		t.Fatalf("expected aps payload to be set")
	}

	if got := cfg.Payload.Aps.Sound; got != "default" {
		t.Errorf("sound = %q, want default", got)
	}
}
