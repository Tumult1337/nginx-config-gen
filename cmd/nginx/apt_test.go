package main

import (
	"strings"
	"testing"
)

// Regression: every apt-get install must pass --force-confdef + --force-confold
// or dpkg blocks on a y/n prompt for any locally-modified conffile (e.g. a
// hand-edited /etc/nginx/nginx.conf on a --convert host) and dies on EOF when
// stdin is closed.
func TestAptInstallPassesDpkgConffileFlags(t *testing.T) {
	d, exec, _, _ := defaultDepsFor(t)
	if _, err := aptInstall(d, "-y", "nginx"); err != nil {
		t.Fatalf("aptInstall: %v", err)
	}
	if len(exec.calls) != 1 {
		t.Fatalf("want 1 call, got %d: %v", len(exec.calls), exec.calls)
	}
	got := exec.calls[0]
	if got[0] != "apt-get" {
		t.Errorf("argv[0] = %q, want apt-get", got[0])
	}
	joined := strings.Join(got[1:], " ")
	for _, want := range []string{
		"-o Dpkg::Options::=--force-confdef",
		"-o Dpkg::Options::=--force-confold",
		"install",
		"-y nginx",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in args: %v", want, got[1:])
		}
	}
	// The dpkg -o flags must come BEFORE the `install` subcommand
	// (apt-get parses -o as a global option, not as install args).
	installIdx := -1
	for i, a := range got {
		if a == "install" {
			installIdx = i
			break
		}
	}
	if installIdx == -1 {
		t.Fatalf("no install in args: %v", got)
	}
	pre := strings.Join(got[1:installIdx], " ")
	if !strings.Contains(pre, "Dpkg::Options::=--force-confdef") ||
		!strings.Contains(pre, "Dpkg::Options::=--force-confold") {
		t.Errorf("dpkg flags must precede 'install', got pre=%q", pre)
	}
}
