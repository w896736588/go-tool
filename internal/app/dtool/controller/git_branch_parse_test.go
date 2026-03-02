package controller

import "testing"

func TestParseBranchFromCurrentBranchOutput_FilterNoise(t *testing.T) {
	output := `
__DT_LOCAL_BRANCH_BEGIN__
1772435109191338300_:%d\n' "$?"
feature/demo_branch_20260302
__DT_LOCAL_BRANCH_END__
__DT_REMOTE_BRANCH_BEGIN__
frog@iZbp18rsv13t3c3a1hzqikZ:/var/www/docker_apps/common/yii_customer_service$
d548695a65f60f9c81768fcd4f34d6f4d5df0ce4	refs/heads/feature/demo_branch_20260302
__DT_REMOTE_BRANCH_END__
`

	local, remote := parseBranchFromCurrentBranchOutput(output)
	if local != "feature/demo_branch_20260302" {
		t.Fatalf("local = %q, want %q", local, "feature/demo_branch_20260302")
	}
	if remote != "d548695a65f60f9c81768fcd4f34d6f4d5df0ce4\trefs/heads/feature/demo_branch_20260302" {
		t.Fatalf("remote = %q, want refs line", remote)
	}
}

func TestBuildCurrentBranchDisplayOutput(t *testing.T) {
	got := buildCurrentBranchDisplayOutput("feature/demo", "master")
	want := "当前分支：\nfeature/demo\n远程分支：\nmaster"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
