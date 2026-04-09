const assert = require('assert')

const MODULE_PATH = '../src/utils/md_editor_config.cjs'

const loadMdEditorConfig = () => require(MODULE_PATH)

const run = () => {
  const {
    buildMdEditorCodeMirrorExtensions,
  } = loadMdEditorConfig()

  const extensions = [
    { type: 'lineWrapping', extension: {} },
    { type: 'markdown', extension: {} },
    { type: 'linkShortener', extension: {}, options: { maxLength: 30 } },
    { type: 'floatingToolbar', extension: {} },
  ]

  const nextExtensions = buildMdEditorCodeMirrorExtensions(extensions)

  assert.deepStrictEqual(
    nextExtensions.map(item => item.type),
    ['lineWrapping', 'markdown', 'floatingToolbar'],
    '知识片段编辑器不应启用 linkShortener，避免 URI 被折叠成省略展示'
  )

  assert.deepStrictEqual(
    extensions.map(item => item.type),
    ['lineWrapping', 'markdown', 'linkShortener', 'floatingToolbar'],
    '过滤逻辑不应直接修改原始扩展数组'
  )

  console.log('md_editor_config tests passed')
}

run()
