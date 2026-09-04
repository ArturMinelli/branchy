# Research: Manual Remote Count Reload

**Feature**: `016-tui-reload-counts` | **Date**: 2026-09-04

## 1. Reuse vs new fetch path

**Decision**: Bind `r` to the existing `remoteUpdateCmd` → `git.FetchDefaultRemote` → `remoteUpdateMsg` → `applyRemoteUpdate` / `SyncFlowModel.applyFileCounts` pipeline from 007.

**Rationale**: Spec requires same default remote, same silent failure policy, same in-place refresh, and coalescing with auto-fetch. Duplicating fetch logic would violate FR-011 and YAGNI.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| New `manualReloadCmd` with different msg type | Same behavior as auto-fetch; doubles apply paths |
| Call `loadInboundCounts` without fetch | Stale remote refs; defeats the feature |
| `git fetch` directly in key handler | Blocks the UI; breaks FR-005 |

---

## 2. In-flight coalescing at TUI vs git

**Decision**: Always schedule `remoteUpdateCmd` on `r`. Do not add a TUI-level in-flight flag.

**Rationale**: `FetchDefaultRemote` already single-flights per absolute repo path (007 research §4). A second cmd waits on the shared flight and delivers a `remoteUpdateMsg` when done — satisfying “coalesce with in-flight update.” No extra state; rapid `r` during fetch cannot spawn parallel network work.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| TUI no-op when fetch in flight | Harder to track (App vs standalone sync); marginal benefit |
| Cancel and restart fetch | Rejected in specify grilling |
| Process-wide debounce timer | Polling-ish; out of spec scope |

---

## 3. Surfaces and key routing

**Decision**:

| Surface | `r` bound | Help shows `r` |
|---------|-----------|----------------|
| Main tree | Yes | Yes (`treeHelpFooter`) |
| Sync root picker | Yes | Yes (`syncPickerHelp`) |
| Sync edge confirm | Yes | No (minimal confirm chrome) |
| Sync summary / browser / processing | No | No |
| MR / link / unlink / projects | No | No |

**Rationale**: Matches specify grilling (tree + sync picker/confirm). FR-008 mandates tree footer only; plan grilling adds picker help for discoverability. Confirm step gets the key but not help — confirm chrome is intentionally minimal today.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Tree only | Sync is P1 in spec; users need reload during sync |
| Help on confirm too | Clutters yes/no panel; key is enough for power users |

---

## 4. Count recompute scope

**Decision**: On successful `remoteUpdateMsg`, recompute **both** inbound and outbound maps via `loadInboundCounts` + `loadOutboundCounts`, then `setFileCounts` / `applyFileCounts` (current 008 behavior in `applyRemoteUpdate`).

**Rationale**: FR-009 — direction toggle must show fresh data immediately after reload without a second fetch.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Inbound only | Outbound badges stale after toggle |
| Recompute only visible direction | Violates FR-009 |

---

## 5. Embedded sync message routing

**Decision**: No routing change. App continues to intercept `remoteUpdateMsg` before delegating to `syncFlow` when `screen == screenSync` (007 research §5). Manual reload from sync picker/confirm in embedded mode: `syncFlow` returns `remoteUpdateCmd` from its key handler; App's top-level `Update` still handles the resulting msg for tree + sync refresh.

**Rationale**: Existing architecture already refreshes both surfaces. Standalone `branchy sync` handles msg inside `SyncFlowModel.Update` directly.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Forward reload only to App | Breaks standalone sync |
| Duplicate apply in sync only | Tree would stay stale when returning from sync |

---

## 6. Testing strategy

**Decision**:

- **Key wiring**: `updateTree` / `updatePickRoot` / `updateEdgeConfirm` with `r` returns non-nil cmd; mock by checking cmd produces `remoteUpdateMsg` (pattern from `remote_test.go`).
- **Refresh**: Continue injecting `remoteUpdateMsg` — no network in unit tests.
- **Help**: Assert footer/picker strings contain `r: reload`.
- **Coalescing**: Covered by existing `TestFetchDefaultRemoteConcurrentShare` in `internal/git`; no new git tests unless behavior changes.

**Rationale**: Feature is wiring; fetch semantics are frozen from 007.

---

## 7. `r` key conflicts

**Decision**: `r` is safe on all bound surfaces — not used by navigation, confirm (y/n), direction chords, or quit.

**Rationale**: Grep of `internal/tui` shows no existing `r` binding. Confirm panel uses y/n/enter only.

**Alternatives considered**: None — no conflict found.
