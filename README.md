# Mulholland skills

Claude Code skills we use every day at [Mulholland](https://mulholland.ai).

## Install

```
/plugin marketplace add Mulholland-Inc/skills
/plugin install scrutinize@mulholland
```

Or copy a skill folder (`plugins/<name>/skills/<name>`) into `~/.claude/skills/`.

## scrutinize

`/scrutinize <paths, a package or a workflow>` — radical simplicity for the scope you bound.

Alone in a git worktree, Claude deletes, merges and cuts until the scope is much
smaller and easier for a fresh agent to work in. Every change is gated by a hard
score — lines of code, how many tool calls a fresh agent needs to answer probes
about the code, and memory and CPU when there is a benchmark — and lands only when
the score drops, no pillar regresses and behavior is proven unchanged. It ends in
a draft PR, where you take back any cut you disagree with.

It never breaks data or anything an outside client calls.

Needs `git`, `go` (the scorer is one stdlib-only Go file) and `gh` for the PR.

## License

MIT
