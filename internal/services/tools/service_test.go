package tools

import "testing"

func TestNew_DefaultThumbSemaphoreCap(t *testing.T) {
	s := New(nil)
	if got := s.thumbSem.Cap(); got != 1 {
		t.Errorf("thumbSem cap: want 1, got %d", got)
	}
}

func TestInstallTool_UnknownToolReturnsError(t *testing.T) {
	s := New(nil)
	if err := s.InstallTool("garbage"); err == nil {
		t.Error("expected error for unknown tool, got nil")
	}
}

func TestDetectBinaries_ReturnsMap(t *testing.T) {
	// DetectBinaries inspects PATH + WinGet locations; in the test
	// environment we cannot guarantee binaries are present, but the
	// call must not panic and must always return a non-nil map.
	s := New(nil)
	got := s.DetectBinaries()
	if got == nil {
		t.Error("DetectBinaries returned nil map")
	}
}
