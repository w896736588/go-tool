const fs = require('fs')
const path = require('path')

const notifyPath = path.join(__dirname, '..', 'src', 'utils', 'base', 'notify.js')
const notifySource = fs.readFileSync(notifyPath, 'utf8')

if (notifySource.includes('Vue.prototype.$notify')) {
  throw new Error('notify.info should not use Vue.prototype.$notify in Vue 3 runtime')
}

if (!/function info\(msg\)\s*\{\s*globals\.\$notify\(/s.test(notifySource)) {
  throw new Error('notify.info should call globals.$notify')
}

console.log('notify info runtime guard passed')
