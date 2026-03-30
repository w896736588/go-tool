const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave : false,
  devServer: {
    proxy: {
      '^/api': {
        target: 'http://127.0.0.1:17170',
        changeOrigin: true,
      },
      '^/sse': {
        target: 'http://127.0.0.1:17170',
        changeOrigin: true,
      },
      '^/socket': {
        target: 'ws://127.0.0.1:17171',
        ws: true,
        changeOrigin: true,
      },
    },
  },
})
