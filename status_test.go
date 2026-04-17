package main

import "testing"

func TestParseStatusClean(t *testing.T) {
	m, s, u := parseStatusPorcelain([]byte(""))
	if m != 0 || s != 0 || u != 0 {
		t.Errorf("clean repo: modified=%d staged=%d untracked=%d, want all 0", m, s, u)
	}
}

func TestParseStatusMixed(t *testing.T) {
	data := []byte(` M file1.go
M  file2.go
A  file3.go
?? newfile.txt
?? another.txt
MM both.go
`)
	m, s, u := parseStatusPorcelain(data)
	// file1.go: modified (y=M)
	// file2.go: staged (x=M)
	// file3.go: staged (x=A)
	// newfile.txt: untracked
	// another.txt: untracked
	// both.go: staged (x=M) + modified (y=M)
	if m != 2 { // file1.go, both.go
		t.Errorf("modified = %d, want 2", m)
	}
	if s != 3 { // file2.go, file3.go, both.go
		t.Errorf("staged = %d, want 3", s)
	}
	if u != 2 { // newfile.txt, another.txt
		t.Errorf("untracked = %d, want 2", u)
	}
}

func TestParseStatusUntrackedOnly(t *testing.T) {
	data := []byte("?? foo.txt\n?? bar.txt\n?? baz.txt\n")
	m, s, u := parseStatusPorcelain(data)
	if m != 0 || s != 0 {
		t.Errorf("modified=%d staged=%d, want 0/0", m, s)
	}
	if u != 3 {
		t.Errorf("untracked = %d, want 3", u)
	}
}

func TestParseStatusStagedOnly(t *testing.T) {
	data := []byte("A  new.go\nM  changed.go\nD  deleted.go\n")
	m, s, u := parseStatusPorcelain(data)
	if s != 3 {
		t.Errorf("staged = %d, want 3", s)
	}
	if m != 0 || u != 0 {
		t.Errorf("modified=%d untracked=%d, want 0/0", m, u)
	}
}

func TestParseStatusModifiedOnly(t *testing.T) {
	data := []byte(" M one.go\n M two.go\n")
	m, s, u := parseStatusPorcelain(data)
	if m != 2 {
		t.Errorf("modified = %d, want 2", m)
	}
	if s != 0 || u != 0 {
		t.Errorf("staged=%d untracked=%d, want 0/0", s, u)
	}
}
