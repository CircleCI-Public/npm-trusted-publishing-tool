package main

import (
	"errors"
	"testing"
)

func TestVCSOrigin(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"https://github.com/acme/widget", "github.com/acme/widget"},
		{"http://github.com/acme/widget", "github.com/acme/widget"},
		{"https://github.com/acme/widget/", "github.com/acme/widget"},
		{"github.com/acme/widget", "github.com/acme/widget"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := vcsOrigin(tt.in); got != tt.want {
			t.Errorf("vcsOrigin(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCircleciError(t *testing.T) {
	tests := []struct {
		name           string
		stderr, stdout []byte
		runErr         error
		want           string
	}{
		{
			name:   "message from stderr json",
			stderr: []byte(`{"error":true,"message":"No CircleCI API token found."}`),
			runErr: errors.New("exit status 3"),
			want:   "No CircleCI API token found.",
		},
		{
			name:   "message with suggestions appended",
			stderr: []byte(`{"message":"No CircleCI API token found.","suggestions":["Run: circleci settings set token <t>","Or set CIRCLE_TOKEN"]}`),
			runErr: errors.New("exit status 3"),
			want:   "No CircleCI API token found. (Run: circleci settings set token <t>; Or set CIRCLE_TOKEN)",
		},
		{
			name:   "message from stdout json when stderr empty",
			stdout: []byte(`{"message":"boom"}`),
			runErr: errors.New("exit status 1"),
			want:   "boom",
		},
		{
			name:   "raw stderr when not json",
			stderr: []byte("  plain error text  "),
			runErr: errors.New("exit status 1"),
			want:   "plain error text",
		},
		{
			name:   "falls back to run error",
			runErr: errors.New("exec: not found"),
			want:   "exec: not found",
		},
		{
			name:   "json without message returns raw stderr",
			stderr: []byte(`{"error":true}`),
			runErr: errors.New("exit status 2"),
			want:   `{"error":true}`,
		},
		{
			name:   "json without message and empty stderr falls back to run error",
			stdout: []byte(`{"error":true}`),
			runErr: errors.New("exit status 2"),
			want:   "exit status 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := circleciError(tt.stderr, tt.stdout, tt.runErr); got != tt.want {
				t.Errorf("circleciError() = %q, want %q", got, tt.want)
			}
		})
	}
}
