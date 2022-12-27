<template>
  <el-card>
    <el-card style="margin-top: 20px;" >
      <div slot="header" class="clearfix">
        <span>测试环境配置</span>
      </div>
      <el-form ref="form" :model="sshConfig" label-width="80px">
        <el-form-item label="账号名">
          <el-input v-model="sshConfig.username"></el-input>
        </el-form-item>
        <el-form-item label="密码">
          <el-input type="password" v-model="sshConfig.password"></el-input>
        </el-form-item>
        <el-form-item label="IP地值">
          <el-input v-model="sshConfig.host"></el-input>
        </el-form-item>
        <el-form-item label="端口">
          <el-input v-model="sshConfig.port"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSshConfig('dev')">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

<!--    <el-card style="margin-top: 20px;" >-->
<!--      <div slot="header" class="clearfix">-->
<!--        <span>预发布环境配置（支持git相关操作）</span>-->
<!--      </div>-->
<!--      <el-form ref="form" :model="prodTestSshConfig" label-width="80px">-->
<!--        <el-form-item label="账号名">-->
<!--          <el-input v-model="prodTestSshConfig.username"></el-input>-->
<!--        </el-form-item>-->
<!--        <el-form-item label="密码">-->
<!--          <el-input type="password" v-model="prodTestSshConfig.password"></el-input>-->
<!--        </el-form-item>-->
<!--        <el-form-item label="IP地值">-->
<!--          <el-input v-model="prodTestSshConfig.host"></el-input>-->
<!--        </el-form-item>-->
<!--        <el-form-item label="端口">-->
<!--          <el-input v-model="prodTestSshConfig.port"></el-input>-->
<!--        </el-form-item>-->
<!--        <el-form-item>-->
<!--          <el-button type="primary" @click="saveSshConfig('prodTest')">保存配置</el-button>-->
<!--        </el-form-item>-->
<!--      </el-form>-->
<!--    </el-card>-->
  </el-card>
</template>

<script>
import {Message} from "element-ui";

export default {
  data() {
    return {
      name: "Ssh",
      //ssh config
      sshConfig: {
        username: "",
        password: "",
        host: "121.40.109.241",
        port: "22",
      },
      //prod ssh config
      prodTestSshConfig: {
        username: "",
        password: "",
        host: "47.96.139.231",
        port: "22",
      },
    }
  },
  mounted: function () {
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig !== null) {
      this.sshConfig = JSON.parse(sshConfig)
    }

    let prodTestSshConfig = this.getStore('prodTestSshConfig')
    if (prodTestSshConfig !== null) {
      this.prodTestSshConfig = JSON.parse(prodTestSshConfig)
    }
  },
  methods: {
    //保存ssh配置
    saveSshConfig: function (prefix) {
      if(prefix === 'prodTest'){
        if (this.prodTestSshConfig.username === '') {
          this.error('用户名不能为空')
          return;
        } else if (this.prodTestSshConfig.password === '') {
          this.error('密码不能为空')
          return;
        } else if (this.prodTestSshConfig.host === '') {
          this.error('连接IP不能为空')
          return;
        } else if (this.prodTestSshConfig.port === '') {
          this.error('端口不能为空')
          return;
        }
        this.setStore('prodTestSshConfig', JSON.stringify(this.prodTestSshConfig))
        this.success('设置成功')
      }else if(prefix === 'dev'){
        if (this.sshConfig.username === '') {
          this.error('用户名不能为空')
          return;
        } else if (this.sshConfig.password === '') {
          this.error('密码不能为空')
          return;
        } else if (this.sshConfig.host === '') {
          this.error('连接IP不能为空')
          return;
        } else if (this.sshConfig.port === '') {
          this.error('端口不能为空')
          return;
        }
        this.setStore('sshConfig', JSON.stringify(this.sshConfig))
        this.success('设置成功')
      }

    },
    success: function (msg) {
      Message.success(msg);
      //this.$notify({title: '提示', message: msg, type: 'success'});
    },
    warning: function (msg) {
      Message.warning(msg);
      //this.$notify({title: '提示', message: msg, type: 'warning'});
    },
    info: function (msg) {
      Message.info(msg);
      //this.$notify({title: '提示', message: msg});
    },
    error: function (msg) {
      Message.error(msg);
      //this.$notify({title: '提示', message: msg, type: 'error'});
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
