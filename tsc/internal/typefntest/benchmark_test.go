package typefntest

import "testing"

// End-to-end, cold-project measurements include parsing/config, binding, library
// setup and checking. They are deliberately not advertised as VM-only throughput.
func BenchmarkTypefnProject(b *testing.B) {
	for _, source := range []string{"native.ts", "execution.ts", "iteration.ts"} {
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
