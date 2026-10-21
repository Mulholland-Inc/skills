---
name: scrutinize
description: Radical simplicity for a scope the user bounds (paths, a package, a workflow) — alone in a worktree, delete, merge and cut until it is much smaller and easier for a fresh agent to work in, each change gated by a hard score with behavior proven unchanged; then a draft PR. Use when the user runs /scrutinize or asks to squeeze, shrink, clean up or radically simplify part of the code base.
---

# Scrutinize

The user sets the boundary; you do everything inside it without talking to
them. Do the work, do not describe it.

Aim for much smaller, not tidier. Delete whole files, merge split-up ideas,
cut features that cost more than they earn, drop edge cases nobody hits, and
remove dead code, unused options, forwarding layers and impossible checks.
When in doubt, delete: the draft PR is where the user takes a cut back.

Never break, even with a cut, inside or outside the scope:

- data: no schema or migration that drops anything, no write that loses
  records
- what outside callers depend on: public APIs, exported symbols, CLI flags,
  config keys, file and wire formats, URLs and webhooks
- security and money: authentication, authorization, access control,
  isolation between users or tenants, anything that charges or pays

## The score

Lower is better: `S = Σ wᵢ·ln(mᵢ / baseᵢ)`, renormalized over the pillars
the scope has (no benchmark: LOC and agent only); the baseline scores 0.

| pillar | measured as | weight | noise |
|---|---|---|---|
| LOC | non-blank, non-comment lines | 50% | |
| agent | probe tool calls | 20% | 10% |
| memory | bytes per op, median of benchmark runs | 15% | 2% |
| CPU | time per op, median of benchmark runs | 15% | 5% |

A change lands only when S drops and no pillar got worse than its noise.
Comments do not count and they help agents: never delete one that says why.
Never score by eye; [score.py](score.py) (Python 3, standard library only)
is the measurement. `$S` is this skill's directory, `$W` a scratch directory
outside the repo (never committed); no benchmark, no `--bench`. Run it from the
repo root:

```bash
python3 $S/score.py --files $W/scope.txt \
  --agent $W/agent.txt --bench $W/bench.txt > $W/base.json          # once
python3 $S/score.py --files $W/scope.txt \
  --agent $W/agent.txt --bench $W/bench.txt --base $W/base.json \
  --prev $W/prev.json                                               # each try
```

`$W/bench.txt` is one line per benchmark run, whatever the language's tool
printed, rewritten as `<bench> <time_per_op> [<bytes_per_op>]`, in the same
units every time.

### Probes

At baseline write three probes to `$W/probes.md`, asked as a user would, by
what the code does, never by its names, each with a checkable answer: **find**
where X is done, **change** to add a small feature (answered as a diff),
**trace** what happens end to end when Z. They are frozen after the baseline.

Run each twice as a fresh, read-only subagent with none of this
conversation's context, all six at once. The prompt is the probe alone,
never scrutinize, the score or your changes. One line per run to
`$W/agent.txt`: `<probe> <tool_calls>`. A wrong answer refuses the change.

## 1. Freeze the boundary

Resolve the scope into `$W/scope.txt`. If it is ambiguous, ask once; that is
the only question you ever ask. Outside code you only follow: update a caller
a cut breaks, never simplify it. Then `EnterWorktree` on `scrutinize/<slug>`
off the default branch (or `git worktree add`); all work happens there.

## 2. Prove behavior

Proofs are scratch files, never committed, in the repo's own language and
test runner:

- **Code**: golden tests pinning today's outputs, errors and edge cases
  included, beside the existing tests (which run too). For each hot path, a
  benchmark run ten times into `$W/bench.txt` when the language has one.
- **Data**: queries, views or reports snapshot their rows on a local database,
  in a stable order, into `$W/`; they must stay byte-identical.
- **UI**: it builds and type-checks, and its screens look and behave the same
  when run locally.

What you cannot prove, leave alone and say so. Then run the probes, score the
baseline and copy it to `$W/prev.json`.

## 3. Loop

One change at a time, big first. Each pass goes down this list and restarts
at the top when something lands:

1. **Cut** unused files, features, options and paths (`git grep`, and real
   data if the repo has a local copy), or used ones when that makes the scope
   much simpler.
2. **Merge** one idea split up, forwarding layers, near-copies.
3. **Rewrite** to a plainer approach: the standard library or what the repo
   already has, simpler data, one pass, work pushed into the database.
4. **Lines**: unreachable branches, unused parameters, impossible checks,
   single-use helpers, allocations in loops, needless copies.

A cut deletes only the golden tests of the feature it cuts; everything else
still passes. Log it under **Cut** in `$W/ledger.md` with what users lose.
Every other change keeps all behavior. No golf: the formatter decides a line.

Gate in order, stopping at the first failure: formatter, build and lint; the
proof (fresh `$W/bench.txt`); the probes (fresh `$W/agent.txt`, every answer
right); the scorer says `accept`. Accepted: `git add <scope paths>` (never
`-A`), commit with a one-line message, copy the metrics to `$W/prev.json`,
add a ledger line (file, change, Δ%). Refused: restore and move on.

**Stop** when a whole pass lands nothing, or the last five tries together
moved S by less than 0.5%.

## 4. Ship a draft

Delete every scratch test, get the repo's full check green (its CI command,
as its CI runs it), push and `gh pr create --draft`, titled by the repo's
convention. Body: every cut and what it takes away first, then the
scope, before/after pillars and S, files deleted and merged, the proof, the
ledger, and what was left alone and why.

## 5. Report

Two or three sentences and the scorecard line, then nothing until the user
answers:

> server/stream: S −0.31 (LOC 792→514, agent −35%, CPU −8%, mem −22%).
> Deleted 3 files, cut the unused replay option. Draft PR #…; say the word.

On approval: `gh pr ready <n> && gh pr merge <n> --squash --auto`; it lands
once the repo's required reviews and checks pass.

Parallel runs need scopes that do not overlap, and no scope's contract may be
another's interior.
