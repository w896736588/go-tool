<template>
  <div>
    <el-menu
      :default-active="menuName"
      class="el-menu-demo"
      mode="horizontal"
      @select="handleSelect"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b">
      <el-menu-item index="Redis">
        Redis缓存
      </el-menu-item>
      <el-menu-item index="Json">
        JSON格式化
      </el-menu-item>
      <el-menu-item index="Unix">
        时间戳
      </el-menu-item>
      <el-menu-item index="Ip">
        IP查询
      </el-menu-item>
      <el-menu-item index="Translate">
        翻译
      </el-menu-item>
      <el-menu-item index="UrlEncode">
        UrlEncode
      </el-menu-item>

      <el-submenu index="Doc">
        <template slot="title">开发文档</template>
        <el-menu-item index="Doc-1">
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-2">
          <a style="color:white;" target="_blank" href="https://work.weixin.qq.com/kf/">微信客服文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-3">
          <a style="color:white;" target="_blank" href="https://element.eleme.cn/#/zh-CN/component/installation">ElementUI</a>
        </el-menu-item>
      </el-submenu>

    </el-menu>
<!--    内容-->
    <Redis v-show="menuName === 'Redis'"></Redis>
    <Json v-show="menuName === 'Json'"></Json>
    <Unix v-show="menuName === 'Unix'"></Unix>
    <Ip v-show="menuName === 'Ip'"></Ip>
    <UrlEncode v-show="menuName === 'UrlEncode'"></UrlEncode>
    <Translate v-show="menuName === 'Translate'"></Translate>
  </div>

</template>

<script>
import Json from "./Json"
import Unix from "./Unix"
import Ip from "./Ip"
import UrlEncode from "./UrlEncode"
import Translate from "./Translate"
import Redis from "./Cache"
export default {
  data () {
    return {
      name: "Redis",
      menuName : "Redis",
    }
  },
  methods: {
    handleSelect(key, keyPath) {
      this.menuName = keyPath[0];
    }
  },
  components : {
    Json,
    Unix,
    Ip,
    UrlEncode,
    Translate,
    Redis,
  },
}
</script>

<style scoped>

</style>
