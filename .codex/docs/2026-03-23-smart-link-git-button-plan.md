# Smart Link Git Button Plan

> **For Codex:** Implement by first adding a reusable base button component, then migrate `smart_link` usages to it, then verify with a frontend build.

**Goal:** Reuse the Git version management button style across all button interactions under `web/src/components/smart_link` through one base component.

**Architecture:** Create a reusable base wrapper around `el-button` that encapsulates the Git page button style and exposes the common Element Plus button API via attribute passthrough. Replace `smart_link` page-level and dialog-level buttons with this wrapper while leaving non-button controls unchanged.

**Tech Stack:** Vue 3, Element Plus, scoped CSS, existing Vue SFC components

---

### Task 1: Create base reusable button

**Files:**
- Create: `web/src/components/base/GitActionButton.vue`

**Steps:**
1. Add a base Vue component under `web/src/components/base`.
2. Wrap `el-button` and preserve parent attributes, events, and icon/default slots.
3. Encode Git page button colors through CSS variables so later pages can reuse one implementation.

### Task 2: Migrate smart_link button usages

**Files:**
- Modify: `web/src/components/smart_link/link_run.vue`
- Modify: `web/src/components/smart_link/LinkConfigEditor.vue`
- Modify: `web/src/components/smart_link/link_process.vue`
- Modify: `web/src/components/smart_link/link_flow.vue`
- Modify: `web/src/components/smart_link/ProcessItemEditor.vue`

**Steps:**
1. Import the new base button component in each `smart_link` file.
2. Replace `el-button` usages only; keep `el-link`, `el-popconfirm`, icons, and other controls unchanged.
3. Remove duplicated local button CSS that becomes obsolete after the shared component is introduced.

### Task 3: Document the rule

**Files:**
- Create: `AGENTS.md`

**Steps:**
1. Add a repository-level rule for frontend work.
2. State that pages needing the Git button visual style should reuse the base button component instead of redefining local button CSS.

### Task 4: Verify

**Files:**
- Verify only

**Steps:**
1. Run the frontend production build from `web`.
2. Confirm the build passes without introducing button-related compile errors.
