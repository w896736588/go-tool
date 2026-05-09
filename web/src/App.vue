<template>
  <div id="app">
    <router-view/>
    <div
      v-if="sseConnectionCount > 0"
      class="sse-connection-indicator"
      :style="{ backgroundColor: sseConnectionColor }"
      :title="'当前 SSE 连接数: ' + sseConnectionCount"
    >
      SSE {{ sseConnectionCount }}
    </div>
  </div>
</template>

<script>
import base from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
const SseConnectionCountId = 'sse_connection_count'
export default {
  name: 'App',
  components: {
  },
  data() {
    return {
      sseConnectionCount: 0,
    }
  },
  computed: {
    sseConnectionColor() {
      const count = this.sseConnectionCount
      if (count >= 5) return '#F56C6C'
      if (count === 4) return '#E6A23C'
      return '#67C23A'
    },
  },
  mounted() {
    base.DisableSaveShortcut()
    let favicon = document.querySelector('link[rel="icon"]')
    if (process.env.NODE_ENV === 'production') {
      favicon.href = './favicon.ico'
    }
    this.registerSseConnectionCount()
  },
  unmounted() {
    sseDistribute.UnRegisterReceive(SseConnectionCountId)
  },
  methods: {
    registerSseConnectionCount() {
      sseDistribute.RegisterReceive(SseConnectionCountId, (data) => {
        if (typeof data === 'number') {
          this.sseConnectionCount = data
        }
      })
    },
  },
}
</script>

<style>
html,
body,
#app {
  /* 根节点统一占满视口，给内部 100% 高度布局提供稳定基线。
     Let root nodes fill the viewport so inner 100% layouts have a reliable baseline. */
  height: 100%;
}

#app {
  font-family: Consolas , Avenir, Helvetica, Arial, sans-serif !important;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

/* 隐藏所有滚动条 */
::-webkit-scrollbar {
  display: none; /* Chrome/Safari */
}

* {
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}

body{
  margin : 0;
}

.sse-connection-indicator {
  position: fixed;
  right: 16px;
  bottom: 16px;
  padding: 4px 12px;
  border-radius: 12px;
  color: #fff;
  font-size: 12px;
  font-weight: bold;
  z-index: 9999;
  pointer-events: none;
  user-select: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  transition: background-color 0.3s;
}
</style>
