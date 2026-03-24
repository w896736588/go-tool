package controller

import "testing"

func TestParseDockerComposeServiceNames(t *testing.T) {
	raw := "\x1b[?2004hservice_api\r\nservice-web\r\ninvalid service\r\nservice_api\r\nservice-web\r\n"

	got := parseDockerComposeServiceNames(raw)

	want := []string{"service_api", "service-web"}
	if len(got) != len(want) {
		t.Fatalf("service count = %d, want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("service[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseDockerImageRows(t *testing.T) {
	raw := "nginx\tlatest\tsha256:abc123\t2 weeks ago\t187MB\n<none>\t<none>\tsha256:def456\t3 days ago\t120MB\nbad-row\n"

	got := parseDockerImageRows(raw)

	if len(got) != 2 {
		t.Fatalf("image count = %d, want 2, got=%v", len(got), got)
	}
	if got[0]["image_ref"] != "nginx:latest" {
		t.Fatalf("first image_ref = %q, want %q", got[0]["image_ref"], "nginx:latest")
	}
	if got[1]["image_ref"] != "sha256:def456" {
		t.Fatalf("dangling image_ref = %q, want %q", got[1]["image_ref"], "sha256:def456")
	}
}

func TestParseDockerContainerRows(t *testing.T) {
	raw := "c1a2b3\tweb-1\tnginx:latest\trunning\tUp 2 hours\nc9d8e7\tworker-1\tnginx:latest\texited\tExited (0) 3 hours ago\nbad-row\n"

	got := parseDockerContainerRows(raw)

	if len(got) != 2 {
		t.Fatalf("container count = %d, want 2, got=%v", len(got), got)
	}
	if got[0]["container_name"] != "web-1" {
		t.Fatalf("first container_name = %q, want %q", got[0]["container_name"], "web-1")
	}
	if got[1]["state"] != "exited" {
		t.Fatalf("second state = %q, want %q", got[1]["state"], "exited")
	}
}

func TestParseDockerSpaceAnalysisRows(t *testing.T) {
	raw := "abc123\t/web\tnginx:latest\trunning\t/var/lib/docker/containers/abc123/abc123-json.log\t2048\t1024\t4096\n" +
		"def456\t/worker\tbusybox:latest\texited\t\t0\t512\t1024\n" +
		"bad-row\n"

	got := parseDockerSpaceAnalysisRows(raw)

	if len(got) != 2 {
		t.Fatalf("space analysis count = %d, want 2, got=%v", len(got), got)
	}
	if got[0]["container_name"] != "web" {
		t.Fatalf("first container_name = %q, want %q", got[0]["container_name"], "web")
	}
	if got[0]["log_bytes"] != "2048" {
		t.Fatalf("first log_bytes = %q, want %q", got[0]["log_bytes"], "2048")
	}
	if got[1]["log_path"] != "" {
		t.Fatalf("second log_path = %q, want empty", got[1]["log_path"])
	}
}

func TestBuildDockerSpaceSummary(t *testing.T) {
	rows := []map[string]string{
		{
			"log_bytes":     "2048",
			"rw_bytes":      "1024",
			"root_fs_bytes": "4096",
		},
		{
			"log_bytes":     "512",
			"rw_bytes":      "256",
			"root_fs_bytes": "1024",
		},
	}

	got := buildDockerSpaceSummary(rows)

	if got["container_count"] != int64(2) {
		t.Fatalf("container_count = %v, want 2", got["container_count"])
	}
	if got["total_log_bytes"] != int64(2560) {
		t.Fatalf("total_log_bytes = %v, want 2560", got["total_log_bytes"])
	}
	if got["total_rw_bytes"] != int64(1280) {
		t.Fatalf("total_rw_bytes = %v, want 1280", got["total_rw_bytes"])
	}
	if got["total_root_fs_bytes"] != int64(5120) {
		t.Fatalf("total_root_fs_bytes = %v, want 5120", got["total_root_fs_bytes"])
	}
}
