<template>
  <div>
    <el-card>
      <el-card>
      <h3 style="display: inline-block;">
        Docker管理
      </h3>
      <div v-for="(valueDocker,k) in dockerList" :key="k" class="text item" style="margin-top:15px;">
        {{ valueDocker.Name }}
        <el-link type="primary" @click="">检查</el-link>
        <el-link type="danger" @click="restartDocker(valueDocker)">重启</el-link>
        <el-link type="primary" @click="showCompose(valueDocker)">查看compose</el-link>
      </div>
    </el-card>
    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
    </el-card>
  </div>
</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

let dockerList = require("../config/dockerList.json")
export default {
  data() {
    return {
      name: "Vip",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      xkfDevDbConfig: {},
      sshConfig: {},
      prodTestSshConfig: {},
      //选中的vip版本
      chooseVipLevel: -1,
      chooseSystemType: -1,
      //过期时间
      expiredTime: '',
      chooseVipUserName: '',
      dockerList: dockerList,
      //总的操作类型
      ExecType: "",
      execResult: "",//操作结果
      redisConfigList: [],
    }
  },
  mounted: function () {
    if (process.env.NODE_ENV === 'production') {
      this.apiHost = '';
    }
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig !== null) {
      this.sshConfig = JSON.parse(sshConfig)
    }
    if (!this.sshConfig || !this.sshConfig.username || this.sshConfig.username === '') {
      this.error("请先配置ssh");
      return
    }
    let xkfDevDbConfig = this.getStore('devTestDbConfig')
    if (xkfDevDbConfig !== null) {
      this.xkfDevDbConfig = JSON.parse(xkfDevDbConfig)
    }
    let wkSshConfig = this.getStore('wkSshConfig')
    if (wkSshConfig !== null) {
      this.wkSshConfig = JSON.parse(wkSshConfig)
    }
    if(!this.wkSshConfig || !this.wkSshConfig.username || this.wkSshConfig.username === ''){
      this.error("请先配置企微ssh");
      return
    }
  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        DockerList: _that.dockerList,
        ExecType: 'check_all_docker_status',
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    getSshConfig: function (value) {
      //根据类型判断
      let chooseSshConfig = this.sshConfig
      if (value.SshName === 'wk') {
        chooseSshConfig = this.wkSshConfig
      }
      return chooseSshConfig
    },
    //重启docker
    restartDocker: function (dockerValue) {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: this.getSshConfig(dockerValue),
        DockerCodeName: dockerValue.Name,
        ExecType: 'restart_docker',
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    //查看compose配置内容
    showCompose: function (dockerValue) {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: this.getSshConfig(dockerValue),
        DockerCodeName: dockerValue.Name,
        ExecType: 'show_compose',
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    redirectLink: function (linkValue) {
      this.execResult = linkValue.link
      window.open(this.execResult, '_blank');
    },
    success: function (msg) {
      // Message.success(msg);
      this.$notify({title: '提示', message: msg, type: 'success', duration: 1000});
    },
    warning: function (msg) {
      // Message.warning(msg);
      this.$notify({title: '提示', message: msg, type: 'warning', duration: 1000});
    },
    info: function (msg) {
      // Message.info(msg);
      //this.$notify({title: '提示', message: msg});
      this.$notify({title: '提示', message: msg, type: 'info', duration: 1000});
    },
    error: function (msg) {
      // Message.error(msg);
      this.$notify({title: '提示', message: msg, type: 'error', duration: 1000});
    },
    setStore: function (key, value) {
      localStorage.setItem(key, value);
    },
    getStore: function (key) {
      return localStorage.getItem(key);
    }
  },
}
</script>

<style scoped>

</style>
