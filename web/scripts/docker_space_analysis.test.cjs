const assert = require('assert')

const MODULE_PATH = '../src/utils/docker_space_analysis.cjs'

const loadSpaceAnalysisModule = () => require(MODULE_PATH)

const run = () => {
  const {
    normalizeDockerSpaceAnalysisRows,
    createDockerSpaceSummary,
  } = loadSpaceAnalysisModule()

  const normalizedRows = normalizeDockerSpaceAnalysisRows([
    {
      container_name: 'worker',
      log_bytes: '16',
      rw_bytes: '64',
      root_fs_bytes: '256',
    },
    {
      container_name: 'api',
      log_bytes: '128',
      rw_bytes: '32',
      root_fs_bytes: '512',
    },
    {
      container_name: 'web',
      log_bytes: '128',
      rw_bytes: '48',
      root_fs_bytes: '1024',
    },
  ])

  assert.deepStrictEqual(
    normalizedRows.map(item => item.container_name),
    ['web', 'api', 'worker']
  )
  assert.strictEqual(normalizedRows[0].log_bytes_value, 128)
  assert.strictEqual(normalizedRows[1].rw_bytes_value, 32)

  const summary = createDockerSpaceSummary({
    container_count: 3,
    total_log_size: '272B',
    total_rw_size: '144B',
    total_root_fs_size: '1.75KB',
    total_combined_rw_log_size: '416B',
  })

  assert.deepStrictEqual(summary, [
    { label: '容器数', value: 3 },
    { label: '日志占用', value: '272B' },
    { label: '可写层占用', value: '144B' },
    { label: 'RootFS 占用', value: '1.75KB' },
    { label: '日志+可写层', value: '416B' },
  ])

  console.log('docker_space_analysis tests passed')
}

run()
