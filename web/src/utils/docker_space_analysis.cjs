const DEFAULT_SPACE_SUMMARY_VALUE = '--'
const DEFAULT_SPACE_NUMBER = 0

// summaryFieldConfigs 定义空间分析汇总卡片展示顺序，避免页面硬编码文案。
const summaryFieldConfigs = [
  { label: '容器数', key: 'container_count', fallback: DEFAULT_SPACE_NUMBER },
  { label: '日志占用', key: 'total_log_size', fallback: DEFAULT_SPACE_SUMMARY_VALUE },
  { label: '可写层占用', key: 'total_rw_size', fallback: DEFAULT_SPACE_SUMMARY_VALUE },
  { label: 'RootFS 占用', key: 'total_root_fs_size', fallback: DEFAULT_SPACE_SUMMARY_VALUE },
  { label: '日志+可写层', key: 'total_combined_rw_log_size', fallback: DEFAULT_SPACE_SUMMARY_VALUE },
]

function toNumber(value) {
  const numberValue = Number(value)
  if (Number.isFinite(numberValue)) {
    return numberValue
  }
  return 0
}

function normalizeDockerSpaceAnalysisRows(rows) {
  const safeRows = Array.isArray(rows) ? rows : []
  return safeRows
    .map((row) => {
      const safeRow = row || {}
      return {
        ...safeRow,
        log_bytes_value: toNumber(safeRow.log_bytes),
        rw_bytes_value: toNumber(safeRow.rw_bytes),
        root_fs_bytes_value: toNumber(safeRow.root_fs_bytes),
      }
    })
    .sort((left, right) => {
      if (right.log_bytes_value !== left.log_bytes_value) {
        return right.log_bytes_value - left.log_bytes_value
      }
      if (right.rw_bytes_value !== left.rw_bytes_value) {
        return right.rw_bytes_value - left.rw_bytes_value
      }
      return String(left.container_name || '').localeCompare(String(right.container_name || ''))
    })
}

function createDockerSpaceSummary(summary) {
  const safeSummary = summary || {}
  return summaryFieldConfigs.map((item) => ({
    label: item.label,
    value: safeSummary[item.key] ?? item.fallback,
  }))
}

module.exports = {
  createDockerSpaceSummary,
  normalizeDockerSpaceAnalysisRows,
}
