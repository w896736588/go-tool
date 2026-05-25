import assert from 'node:assert/strict'
import codexParser from '../src/utils/codex_chat_parser.js'

function testParsesErrorTextField() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'error',
      text: 'Codex CLI 配置解析失败: missing api key',
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'error')
  assert.equal(messages[0].text, 'Codex CLI 配置解析失败: missing api key')
}

function testParsesThreadStartedWithoutModelAsEmptyString() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'thread.started',
      thread_id: 'thread_123',
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'system_init')
  assert.equal(messages[0].text, '会话已创建')
  assert.equal(messages[0].model, '')
  assert.equal(messages[0].sessionId, 'thread_123')
}

function main() {
  testParsesErrorTextField()
  testParsesThreadStartedWithoutModelAsEmptyString()
  console.log('codex_chat_parser tests passed')
}

main()
