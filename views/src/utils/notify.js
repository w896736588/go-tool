import Vue from 'vue'

function success(msg) {
  Vue.prototype.$notify({title: '提示', message: msg, type: 'success', duration: 1000});
}

function warning(msg) {
  Vue.prototype.$notify({title: '提示', message: msg, type: 'warning', duration: 1000});
}
function info(msg) {
  Vue.prototype.$notify({title: '提示', message: msg, type: 'info', duration: 1000});
}

function error(msg) {
  Vue.prototype.$notify({title: '提示', message: msg, type: 'error', duration: 1000});
}

export default {
  success,
  warning,
  info,
  error,
}
