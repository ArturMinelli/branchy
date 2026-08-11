# Research: Lazy Auto-Fetch

**Feature**: `007-lazy-auto-fetch` | **Date**: 2026-08-11

## 1. When the first frame is allowed to wait

**Decision**: Compute inbound counts synchronously from local refs (current 006 behavior), then return a `tea.Cmd` that fetches. `Model.Init` / `selectProject` / `SyncFlowModel.Init` must not call `git fetch`.

**Rationale**: SC-001 forbids adding network wait to startup. `newModel` currently calls `selectProject` before `tea.NewProgram`; fetch cannot run there either. `Init()` is the first safe place to return a cmd after the program starts, and picker-driven `selectProject` returns a cmd from `Update`.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Fetch in `selectProject` before first `View` | Blocks first paint; violates FR-001 |
| Fetch in a goroutine outside Bubble Tea | Breaks the tea.Cmd model; racey UI updates |
| Show a spinner until fetch ends | Blocks perception of “immediate local info”; grilling said silent |

---

## 2. Default remote

**Decision**: Use `origin` if `git remote` lists it; otherwise the first remote name. If there are no remotes, treat as skip (not a user-visible failure).

**Rationale**: Matches grilling (“default remote, typically origin”) and FR-006 (no remote → keep local, no error). Avoids guessing upstream of the current checkout, which may not be a tree root.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `@{upstream}` of HEAD | Wrong remote if the user is on a feature branch |
| Fetch every remote | Rejected in grilling |
| Fetch only missing tree branch names | Rejected in grilling (“lazily” = when, not what) |

---

## 3. Fetch command

**Decision**: `git -C <dir> fetch <remote>` with no extra flags (no `--prune`, no refspec narrowing, no checkout).

**Rationale**: Normal full update of that remote’s tracking refs. `ResolveRef` already prefers local heads then `origin/<name>`, so a successful fetch is enough for `?` → number. Prune/refspec are YAGNI.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `git fetch --all` | Multiple remotes; grilling said default only |
| `git fetch origin <branch>…` | Incomplete; more code; not “normal fetch” |
| `git pull` | Mutates the working tree; spec forbids checkout updates |

---

## 4. Single-flight per project

**Decision**: Dedupe inside `git.FetchDefaultRemote` with a per-absolute-path in-flight wait. Concurrent callers share one `git fetch`. After it finishes, a later call may start a new fetch (e.g. user re-selects the project).

**Rationale**: FR-004 requires tree and sync to share one update. Putting the lock in `internal/git` covers embedded sync (App started fetch) and standalone `branchy sync` without a shared TUI pointer. Waiting on the in-flight fetch (instead of no-op) means the second caller still gets a completion result and can refresh.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| TUI-only flag on `Model` | Standalone sync has no App; easy to double-fetch |
| Process-wide “fetched once forever” | Re-selecting a project after a long session would never refresh |
| Mutex without sharing the result | Second caller would start a second `git fetch` |

---

## 5. Applying the result in the TUI

**Decision**: `remoteUpdateMsg{projectID, path, err}`. App handles it even when `screen == screenSync` (intercept before delegating all msgs to sync). Success → `loadInboundCounts` + `treeView.setInbound` + `syncFlow.applyInbound` if sync is open. Failure or id mismatch → no-op, no View string.

Standalone sync handles the same msg in `SyncFlowModel.Update`: rebuild picker badges and `resetEdgeConfirm` if on the confirm step.

**Rationale**: Today App forwards every msg to sync while embedded sync is active; an unhandled fetch result would be dropped (FR-005 broken). Intercept + apply keeps one flight and refreshes both surfaces.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Only sync handles the msg when open | Tree cache would stay stale when returning from sync |
| Loading panel during fetch | Violates FR-007 / silent grilling |
| Recompute in `View()` | Forbidden by 006 (no git in View) |

---

## 6. Scripted CLI

**Decision**: Do not call `FetchDefaultRemote` from `internal/cli` or `internal/sync`.

**Rationale**: FR-008; 006 already isolated counts to TUI.

---

## 7. Testing strategy

**Decision**:

- **git**: temp repo + `git init --bare` file:// remote; assert tracking refs appear after `FetchDefaultRemote`; no remotes → nil error; bad remote URL → error; two concurrent calls → one fetch.
- **tui**: inject `remoteUpdateMsg` (do not hit the network). Success updates badges/confirm; failure leaves counts and View has no “fetch” error; msg for another project id is ignored.

**Rationale**: Fetch semantics need a real git remote. UI policy must not depend on network in unit tests.
