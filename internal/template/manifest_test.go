package template

import "testing"

func TestApplyDefaultsSetsPreviewPorts(t *testing.T) {
	manifest := &TemplateManifest{}
	manifest.ApplyDefaults()

	if manifest.Preview.NoVNCPort != 6080 {
		t.Fatalf("expected default novnc port 6080, got %d", manifest.Preview.NoVNCPort)
	}
	if manifest.Preview.VNCPort != 5900 {
		t.Fatalf("expected default vnc port 5900, got %d", manifest.Preview.VNCPort)
	}
}
