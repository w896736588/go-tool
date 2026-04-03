const assert = require('assert')

const {
  mergeSavedSmartLinkIntoList,
} = require('../src/utils/smart_link_config_sync.cjs')

const run = () => {
  const currentList = [
    {
      id: 9,
      name: '测试链接',
      links: JSON.stringify([{ label: '生产环境', link: 'https://example.com' }]),
      open_type: 2,
      open_type_new: 2,
      open_num: 0,
      open_num_new: 0,
      combine_type: '4',
      channel: 'chromium',
      linkList: [{ label: '生产环境', link: 'https://example.com', runNum: 3 }],
    },
  ]

  const savedRecord = {
    id: 9,
    name: '测试链接',
    links: JSON.stringify([{ label: '生产环境', link: 'https://example.com' }]),
    open_type: 3,
    open_num: 0,
    combine_type: 4,
    channel: 'chrome',
  }

  const nextList = mergeSavedSmartLinkIntoList(currentList, savedRecord)

  assert.strictEqual(nextList.length, 1, '应保留原列表长度')
  assert.strictEqual(nextList[0].open_type, 3, '保存后列表项应同步最新打开类型')
  assert.strictEqual(nextList[0].open_type_new, 3, '运行态打开类型也应同步为最新值')
  assert.strictEqual(nextList[0].combine_type, '4', '组合类型应被规范化为字符串')
  assert.strictEqual(nextList[0].channel, 'chrome', '浏览器通道应同步')
  assert.deepStrictEqual(
    nextList[0].linkList,
    [{ label: '生产环境', link: 'https://example.com' }],
    'linkList 应重新从保存后的 links 生成，避免沿用旧运行态快照'
  )
}

try {
  run()
  console.log('smart_link_save_sync.test.cjs passed')
} catch (error) {
  console.error(error)
  process.exit(1)
}
