# Research: TUI Native Confirm & Loading

**Feature**: `005-sync-tui-polish` | **Date**: 2026-08-11

## 1. Confirm panel component pattern

**Decision**: Implement a lightweight `ConfirmModel` value type in `internal/tui/confirm.go` with `Update(tea.KeyMsg) (ConfirmModel, ConfirmChoice)` and `View() string`, rendered via lipgloss bordered panel and two styled button labels.

**Rationale**: Charm Bracelet `bubbles` has no dedicated confirm dialog. A ~80-line shared model keeps flows decoupled from layout while matching the existing pattern (each `*flow.go` owns step state, delegates render/input to helpers). Lipgloss borders align with `chrome.go` styles already in use.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Inline lipgloss in each flow `View()` | Duplicates focus logic 7×; violates FR-011 |
| `bubbles/list` with two items | Awkward semantics; overkill for binary choice |
| `bubbles/textinput` or viewport | Wrong abstraction; no native yes/no UX |
| Third-party TUI dialog lib | New dependency; YAGNI |

---

## 2. Confirm interaction (arrow focus + shortcuts)

**Decision**: Horizontal focus between `[ No ]` and `[ Yes ]` via `left`/`right`/`tab`/`shift+tab`; default focus on No; `Enter` activates focused button; `y`/`n`/`Esc` shortcuts map to Yes/No/Esc-as-No per spec grilling.

**Rationale**: Matches user grilling decision. Tab is included because some terminals map differently; left/right is primary affordance for binary toggle.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| y/n only with styled border | User chose arrow-key focus |
| Enter always means Yes | Breaks safe-default No convention |

---

## 3. Loading panel / spinner

**Decision**: Use `github.com/charmbracelet/bubbles/spinner` inside `LoadingModel` in `internal/tui/loading.go`. Expose `Init() tea.Cmd` returning `spinner.Tick`, `Update` forwarding spinner msgs, and `View()` combining spinner glyph + message + optional progress line.

**Rationale**: `bubbles` is already a direct dependency (`go.mod`); spinner is the standard Bubble Tea pattern for async work. Sync already has `stepSyncProcessing` + async `tea.Cmd`; only needs spinner wiring. MR and init currently block in `Update` — must move work to `tea.Cmd` to animate spinner and block input.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Static `...` animation via frame counter | Less polished; spec requires animated indicator |
| `bubbles/progress` bar | Indeterminate network work; no % known |
| Goroutine + channel outside tea.Cmd | Breaks Bubble Tea concurrency model |

---

## 4. Async refactor for MR and init

**Decision**: Mirror sync's `runEdgeCmd` pattern:

- MR: `stepMRLoading` after confirm → `tea.Cmd` calling `mr.Create` → result msg → `stepMRResult` or error
- Init: `stepInitLoading` after confirm → `tea.Cmd` calling `project.Init` → success/error step
- Browser batch: `stepSyncBrowserLoading` / `stepMRBrowserLoading` → `tea.Cmd` calling `sync.OpenURLs` or `browser.Open` → finish

**Rationale**: Spinner requires the event loop to keep ticking during work. Synchronous calls in `Update` freeze the UI. Sync already demonstrates the correct pattern.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Fake spinner before sync call returns | Misleading if call is fast; doesn't help slow calls |
| Keep MR/init synchronous, spinner cosmetic only | UI frozen; violates SC-006 perception goal |

---

## 5. Input blocking during loading

**Decision**: Loading steps omit `tea.KeyMsg` handling in flow `Update` (early return before key switch). Only process `spinner.TickMsg`, window resize, and the completion result msg.

**Rationale**: Matches grilling decision (block all input until complete). Simplest correct approach — no cancel race conditions.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Queue keys for after load | Unexpected behavior; user chose block |
| Esc cancels | Explicitly rejected in grilling |

---

## 6. Scope of loading states

**Decision**: Loading panels only for operations listed in spec FR-006–FR-009: sync edge MR creation, MR wizard creation, browser URL open (sync + MR), init registration. Link/unlink confirm → save is synchronous file I/O (<50ms typical); no loading step unless save exceeds perception threshold — out of spec scope.

**Rationale**: Spec lists explicit flows; link/unlink are not listed for loading. Avoid scope creep.

---

## 7. Embedded tree unlink (`app.go`)

**Decision**: Replace `screenUnlinkConfirm` inline `[y/N]` with embedded `ConfirmModel` field on `AppModel`, same component as `unlinkflow.go`.

**Rationale**: FR-001 explicitly includes embedded tree unlink. Reuse shared confirm; app already has separate screen state — swap View/Update to delegate to confirm helper.

---

## 8. Testing strategy

**Decision**:

- Unit tests on `ConfirmModel`: default focus, arrow navigation, y/n/enter/esc outcomes
- Unit tests on `LoadingModel`: spinner tick updates view string
- Flow tests: extend existing `*_test.go` to assert confirm step view contains button markers (not `[y/N]`), loading step ignores keys until result msg

**Rationale**: Matches 004 pattern (`mrflow_test.go` step transitions). Component tests give fast coverage; flow tests guard integration.

---

## 9. Scripted CLI unchanged

**Decision**: No changes to `internal/sync/sync.go` stdin prompts or `internal/cli` plain paths. TUI-only changes in `internal/tui/*`.

**Rationale**: FR-013; `UseTUI` gate already isolates interactive mode.
