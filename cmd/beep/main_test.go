package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/beep", "version")
	cmd.Dir = projectRoot()
	cmd.Env = append(cmd.Env, "PATH=/usr/local/go/bin:"+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "GOCACHE="+os.Getenv("GOCACHE"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, output)
	}
	if string(output) != "beep v0.1.0\n" {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestUsageOnNoArgs(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/beep")
	cmd.Dir = projectRoot()
	cmd.Env = append(cmd.Env, "PATH=/usr/local/go/bin:"+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "GOCACHE="+os.Getenv("GOCACHE"))
	err := cmd.Run()
	if err == nil {
		t.Error("expected error exit code when no args provided")
	}
}

func TestHelpCommand(t *testing.T) {
	tests := []string{"help", "--help", "-h"}
	for _, arg := range tests {
		cmd := exec.Command("go", "run", "./cmd/beep", arg)
		cmd.Dir = projectRoot()
		cmd.Env = append(cmd.Env, "PATH=/usr/local/go/bin:"+os.Getenv("PATH"))
		cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
		cmd.Env = append(cmd.Env, "GOCACHE="+os.Getenv("GOCACHE"))
		output, _ := cmd.CombinedOutput()
		// help exits with code 1 (usage function calls os.Exit(1))
		if !contains(string(output), "Usage:") {
			t.Errorf("expected help output for %q, got %q", arg, output)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/beep", "unknown")
	cmd.Dir = projectRoot()
	cmd.Env = append(cmd.Env, "PATH=/usr/local/go/bin:"+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "GOCACHE="+os.Getenv("GOCACHE"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error exit code for unknown command")
	}
	if !contains(string(output), "Unknown command") {
		t.Errorf("expected 'Unknown command' in output, got %q", output)
	}
}

func TestServeHelp(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/beep", "serve", "--help")
	cmd.Dir = projectRoot()
	cmd.Env = append(cmd.Env, "PATH=/usr/local/go/bin:"+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "GOCACHE="+os.Getenv("GOCACHE"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("serve --help failed: %v\n%s", err, output)
	}
	if !contains(string(output), "port") {
		t.Errorf("expected 'port' in serve help, got %q", output)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}
