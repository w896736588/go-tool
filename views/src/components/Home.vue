<template>
  <div>

    <el-menu
      :default-active="menuName"
      class="el-menu-demo"
      router
      mode="horizontal"
      @select="handleSelect"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b">
      <el-menu-item index="/CacheIndex">
        Redis
      </el-menu-item>
      <el-menu-item index="/Consumer">
        消费者
      </el-menu-item>
      <el-menu-item index="/Git">
        Git
      </el-menu-item>
      <el-menu-item index="/WechatKefu">
        微信客服
      </el-menu-item>
      <el-menu-item index="/Vip">
        版本变更
      </el-menu-item>
      <el-menu-item index="/Link">
        登录/链接
      </el-menu-item>
      <el-menu-item index="/Docker">
        Docker
      </el-menu-item>
      <el-menu-item index="/Model">
        Model/建表Sql
      </el-menu-item>
      <!--      <el-menu-item index="Tools">-->
      <!--        小工具-->
      <!--      </el-menu-item>-->
      <el-menu-item index="/Ssh">
        服务器配置
      </el-menu-item>

      <el-submenu index="">
        <template slot="title">开发文档</template>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://kf.weixin.qq.com/api/doc/path/93304">微信客服文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://open.work.weixin.qq.com/api/doc/90000/90135/90664">企业微信文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://element.eleme.cn/#/zh-CN/component/installation">ElementUI</a>
        </el-menu-item>
      </el-submenu>
    </el-menu>
    <el-tag style="margin:10px;"
            v-for="tag in tags"
            :key="tag.name"
            closable
            :type="tag.type">
      {{tag.name}}
    </el-tag>
    <el-main>
      <router-view name="home"></router-view>
    </el-main>
  </div>

</template>

<script>
export default {
  data () {
    return {
      menuName : "/CacheIndex",
      tags : [],
    }
  },
  mounted : function (){
    this.menuName = this.$helperStore.getStore('lastMenuName')
    if(!this.$helperConfig.getXkfDevSshConfig() || !this.$helperConfig.getWkDevSshConfig() || !this.$helperConfig.getXkfDevDbConfig()){
      this.menuName = '/Ssh';
    }
    if (this.$route.path !== this.menuName) {
      this.$router.push(this.menuName)
    }
  },
  methods: {
    showNotify : function (notifyList){
      this.tags = notifyList
    },
    handleSelect(key, keyPath) {
      console.log(keyPath , key)
      console.log(keyPath)
      if(keyPath[0].indexOf('Doc-') >= 0){
        return
      }
      this.menuName = keyPath[0];
      this.$helperStore.setStore('lastMenuName' , this.menuName)
    }
  },
  components : {

  },
}
</script>

<style scoped>

</style>
