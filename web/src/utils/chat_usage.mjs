export function formatInlineUsage(usage) {
  const data = usage || {}
  const inputTokens = data.input_tokens ?? 0
  const outputTokens = data.output_tokens ?? 0
  const cacheReadInputTokens = data.cache_read_input_tokens ?? 0
  const cacheCreationInputTokens = data.cache_creation_input_tokens ?? 0

  return `input: ${inputTokens} | output: ${outputTokens} | cache: ${cacheReadInputTokens} | cache create: ${cacheCreationInputTokens}`
}
