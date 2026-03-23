const assert = require('assert')

const MODULE_PATH = '../src/utils/home_dashboard_wheel.cjs'

const loadWheelModule = () => require(MODULE_PATH)

const createScrollableElement = ({
  clientHeight,
  scrollHeight,
  scrollTop,
  parentElement = null,
}) => ({
  clientHeight,
  scrollHeight,
  scrollTop,
  parentElement,
})

const run = () => {
  const { shouldBlockHomeDashboardPageSwitch } = loadWheelModule()

  const rootElement = createScrollableElement({
    clientHeight: 480,
    scrollHeight: 480,
    scrollTop: 0,
  })
  const processContainer = createScrollableElement({
    clientHeight: 200,
    scrollHeight: 600,
    scrollTop: 120,
    parentElement: rootElement,
  })
  const processTextLine = {
    parentElement: processContainer,
  }

  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, 48),
    true,
    '内部输出区域还能继续向下滚动时，不应该触发首页翻页'
  )

  processContainer.scrollTop = 80
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, -48),
    true,
    '内部输出区域还能继续向上滚动时，不应该触发首页翻页'
  )

  processContainer.scrollTop = 0
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, -48),
    false,
    '内部输出区域已经滚到顶部时，允许继续处理向上翻页'
  )

  processContainer.scrollTop = 400
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, 48),
    false,
    '内部输出区域已经滚到底部时，允许继续处理向下翻页'
  )

  const staticContainer = createScrollableElement({
    clientHeight: 240,
    scrollHeight: 240,
    scrollTop: 0,
    parentElement: rootElement,
  })
  const staticChild = {
    parentElement: staticContainer,
  }
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(staticChild, 48),
    true,
    '非可滚动区域滚轮事件不应该触发首页翻页'
  )

  console.log('home_dashboard_wheel tests passed')
}

run()
