import assert from 'node:assert/strict'
import chatParser from '../src/utils/chat_parser.js'

function testParsesStandaloneInputJsonDelta() {
  const messages = chatParser.parseChatLines([
    JSON.stringify({
      type: 'stream_event',
      event: {
        type: 'content_block_delta',
        delta: {
          type: 'input_json_delta',
          partial_json: '"',
        },
      },
    }),
    JSON.stringify({
      type: 'stream_event',
      event: {
        type: 'content_block_delta',
        delta: {
          type: 'input_json_delta',
          partial_json: 'content',
        },
      },
    }),
    JSON.stringify({
      type: 'stream_event',
      event: {
        type: 'content_block_delta',
        delta: {
          type: 'input_json_delta',
          partial_json: '"',
        },
      },
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'assistant')
  assert.equal(messages[0].content.length, 1)
  assert.equal(messages[0].content[0].type, 'tool_use')
  assert.equal(messages[0].content[0].input, '"content"')
}

testParsesStandaloneInputJsonDelta()
console.log('chat_parser stream_event tests passed')
