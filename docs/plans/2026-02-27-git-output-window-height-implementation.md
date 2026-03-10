# Git Output Window Height Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the `/Git` output window stick to the bottom and automatically fill remaining height.

**Architecture:** Use CSS flex layout for the Git page container and output card, and update the shell result component to optionally use container height instead of a fixed `divHeight` calculation.

**Tech Stack:** Vue 3, Element Plus, CSS flexbox.

---

### Task 1: Add container-height mode to shell result component

**Files:**
- Modify: `web/src/components/shell/result_div.vue`

**Step 1: Write the failing test**

No automated UI test framework exists in this repo. Document manual verification instead.

**Step 2: Run test to verify it fails**

Skip (no automated test runner for UI layout).

**Step 3: Write minimal implementation**

- Add a boolean prop such as `useContainerHeight`.
- When `useContainerHeight` is true, set the scrollbar and inner div to `height: 100%` (no `divHeight` math).
- Preserve existing behavior when the new prop is false.

**Step 4: Run test to verify it passes**

Manual check on `/Git` after Task 2 (see Task 3).

**Step 5: Commit**

```bash
git add web/src/components/shell/result_div.vue
git commit -m "feat: add container-height mode for shell output"
```

### Task 2: Use container-height mode on `/Git`

**Files:**
- Modify: `web/src/components/Git.vue`

**Step 1: Write the failing test**

No automated UI test framework exists in this repo. Document manual verification instead.

**Step 2: Run test to verify it fails**

Skip (no automated test runner for UI layout).

**Step 3: Write minimal implementation**

- Pass the new `useContainerHeight` prop to `shellResult` in `/Git`.
- Ensure `output-card` and `output-content` remain `flex: 1` with `min-height: 0`.
- If needed, ensure `.git-page-container` can grow to full viewport height (verify parent container constraints first).

**Step 4: Run test to verify it passes**

Manual:
1. Open `/Git` and confirm the output window touches the bottom of the viewport.
2. Resize the browser window; output height should expand and shrink smoothly.

**Step 5: Commit**

```bash
git add web/src/components/Git.vue
git commit -m "fix: make Git output window fill remaining height"
```

### Task 3: Manual verification checklist

**Files:**
- None

**Step 1: Manual checks**

1. `/Git` shows no blank space under the output window at common sizes.
2. Output scroll area remains usable and auto-scroll behavior is unchanged.

**Step 2: Optional lint**

Run: `cd web && npm run lint`
Expected: No new lint errors.

