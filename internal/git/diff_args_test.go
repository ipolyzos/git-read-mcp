package git

import (
	"testing"
)

func intPtr(v int) *int {
	return &v
}

func TestBuildDiffNoIndexArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		left    string
		right   string
		opts    DiffNoIndexOptions
		want    []string
		wantErr bool
	}{
		{
			name:  "defaults",
			left:  "/work/left",
			right: "/work/right",
			opts:  DiffNoIndexOptions{},
			want: []string{
				"diff", "--no-index", "--unified=3",
				"--", "/work/left", "/work/right",
			},
		},
		{
			name:  "unified_zero",
			left:  "/a",
			right: "/b",
			opts: DiffNoIndexOptions{
				Unified: intPtr(0),
			},
			want: []string{
				"diff", "--no-index", "--unified=0",
				"--", "/a", "/b",
			},
		},
		{
			name:  "custom_unified",
			left:  "/a",
			right: "/b",
			opts: DiffNoIndexOptions{
				Unified: intPtr(7),
			},
			want: []string{
				"diff", "--no-index", "--unified=7",
				"--", "/a", "/b",
			},
		},
		{
			name:  "ignore_all_space",
			left:  "/a",
			right: "/b",
			opts: DiffNoIndexOptions{
				Unified: intPtr(5),
				Ignore:  IgnoreWhitespaceAll,
			},
			want: []string{
				"diff", "--no-index", "--unified=5",
				"--ignore-all-space",
				"--", "/a", "/b",
			},
		},
		{
			name:  "ignore_space_change",
			left:  "/a",
			right: "/b",
			opts: DiffNoIndexOptions{
				Ignore: IgnoreWhitespaceChange,
			},
			want: []string{
				"diff", "--no-index", "--unified=3",
				"--ignore-space-change",
				"--", "/a", "/b",
			},
		},
		{
			name:    "unified_too_large",
			left:    "/a",
			right:   "/b",
			opts:    DiffNoIndexOptions{Unified: intPtr(500)},
			wantErr: true,
		},
		{
			name:    "unified_negative",
			left:    "/a",
			right:   "/b",
			opts:    DiffNoIndexOptions{Unified: intPtr(-1)},
			wantErr: true,
		},
		{
			name:    "ignore_invalid",
			left:    "/a",
			right:   "/b",
			opts:    DiffNoIndexOptions{Ignore: IgnoreWhitespace(99)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := buildDiffNoIndexArgs(tt.left, tt.right, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len got=%d want=%d\ngot: %#v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("idx %d: got %q want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestValidatePathArg_blocksFlagLike(t *testing.T) {
	t.Parallel()
	if err := validatePathArg("left", "-not-a-path"); err == nil {
		t.Fatal("expected error")
	}
	if err := validatePathArg("right", ""); err == nil {
		t.Fatal("expected error")
	}
}
