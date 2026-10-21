// Command score measures a frozen scope for /scrutinize and, given a baseline,
// says the scope's score S (lower is better) and whether a change may land.
//
//	go run <skill>/score/main.go -files scope.txt -agent agent.txt [-bench bench.txt] > m.json
//	go run <skill>/score/main.go -files scope.txt -agent agent.txt [-bench bench.txt] -base base.json -prev prev.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The four pillars. Memory and CPU fall away, and the rest renormalize, when the scope has no benchmark.
var weights = map[string]float64{"loc": .50, "agent": .20, "mem": .15, "cpu": .15}

// noise is how much worse a pillar may measure before the change is refused anyway.
var noise = map[string]float64{"agent": .10, "cpu": .05, "mem": .02}

type metrics map[string]float64

func main() {
	files := flag.String("files", "", "file listing the scope, one path per line")
	agent := flag.String("agent", "", "probe runs, one per line: <probe> <tool_calls>")
	bench := flag.String("bench", "", "go test -bench -benchmem -count=N output")
	base := flag.String("base", "", "baseline metrics JSON")
	prev := flag.String("prev", "", "last accepted metrics JSON")
	flag.Parse()

	m := metrics{}
	for _, f := range lines(*files) {
		for _, l := range lines(f) {
			if !comment.MatchString(l) {
				m["loc"]++ // non-blank, non-comment lines
			}
		}
	}
	if *agent != "" {
		probes(*agent, m)
	}
	if *bench != "" {
		benchmarks(*bench, m)
	}
	if *base == "" {
		json.NewEncoder(os.Stdout).Encode(m)
		return
	}
	b, p := load(*base), load(*prev)
	s, sp := score(m, b), score(p, b)
	verdict := "accept"
	if s >= sp {
		verdict = "reject: score did not drop"
	}
	for _, k := range []string{"loc", "agent", "mem", "cpu"} {
		if v, pv := m[k], p[k]; pv > 0 && (v-pv)/pv > noise[k] {
			verdict = fmt.Sprintf("reject: %s worse by %.1f%%", k, 100*(v-pv)/pv)
		}
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"metrics": m, "score": s, "prev": sp, "delta_pct": 100 * (math.Exp(s-sp) - 1), "verdict": verdict,
	})
}

// score is Σ wᵢ·ln(mᵢ/baseᵢ) over the pillars both sides have, renormalized; the baseline scores 0.
func score(m, base metrics) float64 {
	var s, w float64
	for k, wk := range weights {
		if base[k] > 0 && m[k] > 0 {
			s += wk * math.Log(m[k]/base[k])
			w += wk
		}
	}
	if w == 0 {
		return 0
	}
	return s / w
}

var comment = regexp.MustCompile(`^(//|--|#|/\*|\*|<!--)`)

// probes sums each probe's median tool calls over its runs.
func probes(path string, m metrics) {
	runs := map[string][]float64{}
	for _, l := range lines(path) {
		f := strings.Fields(l)
		if len(f) != 2 {
			fail(fmt.Errorf("probe line %q: want <probe> <tool_calls>", l))
		}
		v, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			fail(err)
		}
		runs[f[0]] = append(runs[f[0]], v)
	}
	for _, vs := range runs {
		m["agent"] += median(vs) + 1 // +1 keeps a zero-call probe on the log scale
	}
}

var benchLine = regexp.MustCompile(`^(Benchmark\S+)\s+\d+\s+([\d.]+) ns/op(?:\s+([\d.]+) B/op)?`)

// benchmarks sums each benchmark's median over its -count runs: ns/op is CPU, B/op is memory.
func benchmarks(path string, m metrics) {
	runs := map[[2]string][]float64{} // {benchmark, pillar}
	for _, l := range lines(path) {
		g := benchLine.FindStringSubmatch(l)
		if g == nil {
			continue
		}
		for i, k := range []string{"cpu", "mem"} {
			if v, err := strconv.ParseFloat(g[i+2], 64); err == nil {
				runs[[2]string{g[1], k}] = append(runs[[2]string{g[1], k}], v)
			}
		}
	}
	for k, vs := range runs {
		m[k[1]] += median(vs) + 1 // +1 keeps a zero-byte benchmark on the log scale
	}
}

func median(vs []float64) float64 {
	sort.Float64s(vs)
	return vs[len(vs)/2]
}

func load(path string) metrics {
	m := metrics{}
	b, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &m)
	}
	if err != nil {
		fail(err)
	}
	return m
}

// lines returns a file's non-blank lines, trimmed.
func lines(path string) []string {
	src, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	var out []string
	for _, l := range strings.Split(string(src), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
