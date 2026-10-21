#!/usr/bin/env python3
"""Measure a frozen scope for /scrutinize and, given a baseline, say its score S
(lower is better) and whether a change may land. Standard library only.

    python3 <skill>/score.py --files scope.txt --agent agent.txt [--bench bench.txt] > m.json
    python3 <skill>/score.py --files scope.txt --agent agent.txt [--bench bench.txt] --base base.json --prev prev.json
"""
import argparse, json, math, re, sys

# The four pillars. Memory and CPU fall away, and the rest renormalize, when the scope has no benchmark.
WEIGHTS = {"loc": .50, "agent": .20, "mem": .15, "cpu": .15}

# How much worse a pillar may measure before the change is refused anyway.
NOISE = {"loc": 0, "agent": .10, "cpu": .05, "mem": .02}

COMMENT = re.compile(r"^(//|--|#|/\*|\*|<!--|;|%|\"\"\"|''')")


def lines(path):
    """A file's non-blank lines, trimmed."""
    with open(path, encoding="utf-8", errors="replace") as f:
        return [l.strip() for l in f if l.strip()]


def median(vs):
    return sorted(vs)[len(vs) // 2]


def runs(path, want):
    """Groups each line's numbers by its first field: <name> <value> [<value>]."""
    out = {}
    for l in lines(path):
        f = l.split()
        if not 2 <= len(f) <= want + 1:
            sys.exit(f"{path}: line {l!r}: want <name> and {want} value(s) at most")
        out.setdefault(f[0], []).append([float(v) for v in f[1:]])
    return out


def score(m, base):
    """Σ wᵢ·ln(mᵢ/baseᵢ) over the pillars both sides have, renormalized; the baseline scores 0."""
    s = w = 0.0
    for k, wk in WEIGHTS.items():
        if base.get(k, 0) > 0 and m.get(k, 0) > 0:
            s += wk * math.log(m[k] / base[k])
            w += wk
    return s / w if w else 0.0


def main():
    a = argparse.ArgumentParser()
    a.add_argument("--files", required=True, help="the scope, one path per line")
    a.add_argument("--agent", help="probe runs, one per line: <probe> <tool_calls>")
    a.add_argument("--bench", help="benchmark runs, one per line: <bench> <time_per_op> [<bytes_per_op>]")
    a.add_argument("--base", help="baseline metrics JSON")
    a.add_argument("--prev", help="last accepted metrics JSON")
    o = a.parse_args()

    m = {"loc": float(sum(1 for f in lines(o.files) for l in lines(f) if not COMMENT.match(l)))}
    if o.agent:
        # +1 keeps a zero-call probe on the log scale
        m["agent"] = sum(median([r[0] for r in rs]) + 1 for rs in runs(o.agent, 1).values())
    if o.bench:
        for name, rs in runs(o.bench, 2).items():
            for i, k in enumerate(["cpu", "mem"]):
                vs = [r[i] for r in rs if len(r) > i]
                if vs:
                    m[k] = m.get(k, 0) + median(vs) + 1
    if not o.base:
        json.dump(m, sys.stdout)
        return

    base, prev = (json.load(open(p)) for p in (o.base, o.prev))
    s, sp = score(m, base), score(prev, base)
    verdict = "accept" if s < sp else "reject: score did not drop"
    for k in WEIGHTS:
        v, pv = m.get(k, 0), prev.get(k, 0)
        if pv > 0 and (v - pv) / pv > NOISE[k]:
            verdict = f"reject: {k} worse by {100 * (v - pv) / pv:.1f}%"
    json.dump({"metrics": m, "score": s, "prev": sp,
               "delta_pct": 100 * (math.exp(s - sp) - 1), "verdict": verdict}, sys.stdout)


if __name__ == "__main__":
    main()
