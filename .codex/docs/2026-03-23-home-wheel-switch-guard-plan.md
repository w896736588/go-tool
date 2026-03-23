# 首页滚轮切页保护实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复首页执行命令输出区域滚动时误切换到任务清单页的问题，只在非可滚动区域滚轮时响应整屏切换。

**Architecture:** 将首页滚轮切页判断抽成独立纯函数，专门判断事件目标是否位于当前滚动方向上仍可继续滚动的子容器内。`Home.vue` 继续保留双屏切页入口，只在纯函数判定允许时才执行翻页，从而避免影响 `Dashboard.vue` 内的消息列表、执行过程输出等滚动区域。

**Tech Stack:** Vue 3、Node.js 原生 `assert` 回归脚本、现有首页双屏切页逻辑

---

### Task 1: 补充滚轮保护回归测试

**Files:**
- Create: `web/src/utils/home_dashboard_wheel.cjs`
- Create: `web/scripts/home_dashboard_wheel.test.cjs`

**Step 1: Write the failing test**

在 `web/scripts/home_dashboard_wheel.test.cjs` 中覆盖以下场景：
- 事件目标位于可继续向下滚动的内部容器时，不允许切页
- 事件目标位于可继续向上滚动的内部容器时，不允许切页
- 事件目标不在可滚动容器中时，允许切页

**Step 2: Run test to verify it fails**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: FAIL，提示缺少滚轮保护实现。

**Step 3: Write minimal implementation**

在 `web/src/utils/home_dashboard_wheel.cjs` 中实现纯函数，输入事件目标和滚动方向，返回是否应阻止整屏切页。

**Step 4: Run test to verify it passes**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS

### Task 2: 接入首页滚轮切页逻辑

**Files:**
- Modify: `web/src/components/Home.vue`
- Modify: `web/src/utils/home_dashboard_wheel.cjs`

**Step 1: Integrate guard**

在 `Home.vue` 中引入纯函数，在 `handleDashboardWheel` 的第一页和第二页切页判断前，先根据当前 `event.target` 和滚动方向做保护判断。

**Step 2: Keep existing behavior**

保留原有阈值、动画锁和任务面板边界判断，不调整其它首页交互。

**Step 3: Manual regression notes**

确认首页第一页在命令输出区滚动时不再切到任务清单页，空白区域滚动仍可切页。

### Task 3: 验证

**Files:**
- Test: `web/scripts/home_dashboard_wheel.test.cjs`
- Test: `web/package.json`

**Step 1: Run regression script**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS

**Step 2: Run build verification**

Run: `npm run prod`
Workdir: `web`
Expected: 构建成功，未引入编译错误。
