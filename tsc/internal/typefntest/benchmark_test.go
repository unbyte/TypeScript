package typefntest

import (
	"encoding/json"
	"os"
	"testing"
)

// End-to-end, cold-project measurements include parsing/config, binding, library
// setup and checking. They are deliberately not advertised as VM-only throughput.
func BenchmarkTypefnProject(b *testing.B) {
	for _, source := range []string{"native.ts", "execution.ts", "iteration.ts", "recursive-independent-roots.ts", "recursive-mutual-roots.ts"} {
		b.Run(source, func(b *testing.B) {
			files := map[string]string{"main.ts": source}
			if source == "native.ts" {
				files["helper.ts"] = "native-helper.ts"
			}
			f := fixture{Files: files, Options: map[string]any{"typefn": map[string]any{"version": 1}}}
			b.ReportAllocs()
			for b.Loop() {
				runFixture(b, f)
			}
		})
	}
}

// Includes both the DSL scanner and its native comparison, and all cold-project
// setup. This measures a verified workload, not isolated VM throughput or speedup.
func BenchmarkTypefnCommunityWords(b *testing.B) {
	benchmarkTypefnFixture(b, "community-words")
}

// Includes native comparison and generic call checking as well as DSL traversal.
func BenchmarkTypefnCommunityPaths(b *testing.B) {
	benchmarkTypefnFixture(b, "community-paths")
}

func BenchmarkTypefnCommunityArrayPaths(b *testing.B) {
	benchmarkTypefnFixture(b, "community-array-paths")
}

func benchmarkTypefnFixture(b *testing.B, name string) {
	b.Helper()
	data, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		b.Fatal(err)
	}
	var f fixture
	if err := json.Unmarshal(data, &f); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		runFixture(b, f)
	}
}
