package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func beepCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./cmd/beep-analytics"}, args...)...)
	cmd.Dir = projectRoot()
	cmd.Env = append(cmd.Env,
		"PATH=/usr/local/go/bin:"+os.Getenv("PATH"),
		"HOME="+os.Getenv("HOME"),
		"GOCACHE="+os.Getenv("GOCACHE"),
		"BEEP_SERVER=",
		"BEEP_TOKEN=",
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestVersionCommand(t *testing.T) {
	out, err := beepCmd(t, "version")
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	if out != "beep-analytics v0.1.0\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestUsageOnNoArgs(t *testing.T) {
	_, err := beepCmd(t)
	if err == nil {
		t.Error("expected error exit code when no args provided")
	}
}

func TestHelpCommand(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		out, _ := beepCmd(t, arg)
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output for %q, got %q", arg, out)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	out, err := beepCmd(t, "unknown")
	if err == nil {
		t.Error("expected error exit code for unknown command")
	}
	if !strings.Contains(out, "Unknown command") {
		t.Errorf("expected 'Unknown command' in output, got %q", out)
	}
}

func TestCommandHelp(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"serve", "beep-analytics serve [--port"},
		{"add-site", "beep-analytics add-site <domain>"},
		{"remove-site", "beep-analytics remove-site <domain>"},
		{"list-sites", "beep-analytics list-sites [--server"},
		{"ignore-ip", "beep-analytics ignore-ip <ip>"},
		{"unignore-ip", "beep-analytics unignore-ip <ip>"},
		{"list-ignored", "beep-analytics list-ignored [--server"},
		{"generate-token", "beep-analytics generate-token [--server"},
		{"revoke-token", "beep-analytics revoke-token <id>"},
		{"stats", "beep-analytics stats [--site"},
		{"version", "beep-analytics version"},
	}
	for _, tt := range tests {
		out, err := beepCmd(t, tt.command, "--help")
		if err != nil {
			t.Errorf("%s --help: unexpected exit code %v\n%s", tt.command, err, out)
		}
		if !strings.Contains(out, tt.want) {
			t.Errorf("%s --help: expected %q in output, got %q", tt.command, tt.want, out)
		}
	}
}

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}
