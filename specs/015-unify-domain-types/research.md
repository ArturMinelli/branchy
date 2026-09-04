# Research: Unify Domain Types

**Feature**: `015-unify-domain-types` | **Date**: 2026-08-24

## 1. Canonical merge direction

**Decision**: Keep `sync.Direction` with `Downward` (iota 0, parent → child) and `Upward` (child → parent). Delete TUI `diffDirection`, `diffInbound`, `diffOutbound`, and `syncDirection()`. Rename the main-tree field `app.diffDirection` → `direction sync.Direction`. Treeview, help footers, and the standalone sync picker all store `sync.Direction`.

Mapping that used to live in `syncDirection` becomes a direct comparison in TUI:

| `sync.Direction` | Counts | Arrow | User-facing help (unchanged copy) |
|------------------|--------|-------|-----------------------------------|
| `Downward` (0) | inbound (`git.InboundFiles`) | `↓` | “inbound (parent→child)” / “downward (parent→child)” |
| `Upward` | outbound (`git.OutboundFiles`) | `↑` | “outbound (child→parent)” / “upward (child→parent)” |

Both word pairs stay in help (tree vs sync picker). Scripted `sync` still omits `Direction` (zero value = `Downward`).

`treeview.go` and `inbound.go` import `internal/sync`. `sync` does not import `tui` — no cycle.

**Rationale**: Grilling locked “keep `sync.Direction`; TUI drops the private type.” FR-001/002: one concept, no conversion at the sync boundary. Zero values already align (`diffInbound` and `Downward` are both 0), so session default stays inbound/downward without constructor rewrites.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Rename constants to Inbound/Outbound | Grilling kept Downward/Upward; both user-facing names already appear in help |
| Move `Direction` onto `tree` | Grilling; tree document does not persist direction; `Ends`/`EdgesBelow` already own it |
| Type alias `type diffDirection = sync.Direction` | Still a second name; FR-002 wants no parallel type |
| Unify `git.InboundFiles` / `OutboundFiles` names | Different concept (file-count comparison); spec does not ask for it |

---

## 2. Typed action + skip reason

**Decision**: In `internal/mr`:

```go
type Action string

const (
    ActionCreated Action = "created"
    ActionSkipped Action = "skipped"
    ActionFailed  Action = "failed"
)

type SkipReason string

const (
    SkipNone         SkipReason = ""
    SkipAlreadyOpen  SkipReason = "already_open"
    SkipUserDeclined SkipReason = "user_declined"
)
```

`CreateResult.Action` becomes `Action`. `sync.Result.Action` becomes `mr.Action`. Both results gain `SkipReason mr.SkipReason`.

String values of `Action` stay `"created"` / `"skipped"` / `"failed"` so TUI lines that interpolate `r.Action` do not change.

Setters:

| Event | Action | SkipReason | Message (unchanged) |
|-------|--------|------------|---------------------|
| Created | `ActionCreated` | `SkipNone` | empty |
| Open MR already exists | `ActionSkipped` | `SkipAlreadyOpen` | `"open MR already exists"` |
| User declined confirm | `ActionSkipped` | `SkipUserDeclined` | `"skipped by user"` |
| Unrecovered create failure | `ActionFailed` | `SkipNone` | error text |
| Confirm callback error (sync) | `ActionFailed` | `SkipNone` | error text |

`OpenableURLs` (single filter, SC-005):

- Skip if `URL == ""`
- Include `ActionCreated`
- Include `ActionSkipped` unless `SkipReason == SkipUserDeclined`
- Do **not** inspect `Message`

`processEdge` copies `mrRes.Action`, `URL`, `Message`, and `SkipReason`. `SkippedByUser` sets `SkipUserDeclined`. CLI sync switches on `mr.ActionCreated` etc., not `"created"` literals.

**Rationale**: Grilling locked typed action plus a skip-reason field. Spec entities: created/skipped/failed plus skip subtypes so browser-open filtering has one table. Message remains user-visible copy (outcome freeze).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Keep `Message == "skipped by user"` as discriminator | Grilling; SC-005 fails if a second printer typos the string |
| Extra Action values (`skipped_open` / `skipped_user`) | Grilling; FR-003 keeps three outcomes; user-visible word is still “skipped” |
| Shared printer package | YAGNI; surfaces only need to switch on `mr.Action` |

---

## 3. Create error policy (lock existing split)

**Decision**: Do not change `mr.Create` control flow. Document and test it as the one policy:

| Class | `Create` returns | Surfaces |
|-------|------------------|----------|
| Validation (empty, same ends, unknown branch) | `nil, err` | All report the error; no GitLab call |
| Auth failure | `nil, err` | Same as today (`glab auth: …`) |
| Already-open (find or recover) | `result{skipped, SkipAlreadyOpen}, nil` | Skipped with URL; scripted `mr` exit 0 |
| Unrecovered GitLab create failure | `result{failed}, nil` | TUI shows failed; scripted `mr` maps to `fmt.Errorf` → non-zero; sync records failed and **continues** |

Sync `Run`/`Begin` still abort the **run** on missing `--from` / unknown from-branch / auth before the loop. Per-edge `Create` errors (should be rare after validation) stay mapped to a failed `Result` as today so the cascade does not change.

**Rationale**: Grilling locked “keep failed as a result; CLI `mr` maps to exit.” FR-005 and US3 already describe this split; the code mostly implements it. The leftover mix is `sync.Result.Action string` vs `mr` constants, and CLI sync switching on literals — fixed in research item 2.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| GitLab failure returns `error` from `Create` | Grilling; TUI and sync would lose a uniform result; CLI already maps failed → exit |
| Never return error (validation as failed result) | Grilling; FR-005: validation aborts with an error and must not call GitLab |
| GitLab interface / fake client for US3 | Spec assumption: no GitLab interface |

---

## 4. What this spec does not change

**Decision**: No new packages. No GitLab port. No direction flags or key changes. File-change helpers stay `git.InboundFiles` / `OutboundFiles`. Browser open stays `browser.Open` / `OpenURLs` (014). `WalkDisplay` stays unstyled (014). Scripted sync remains downward-only (no `--direction`).

**Rationale**: FR-006–007, spec assumptions, last of five architecture specs.
