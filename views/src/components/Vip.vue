<template>
  <el-card>

    <el-card class="box-card">
      <div>
        <h3 style="display: inline-block;">
          系统类别
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(value,key) in systemTypeList" style="margin:5px;">
            <div>
              <el-radio size="medium " v-model="chooseSystemType" @change="queryVipInfo" :label="value.level">
                {{ value.name }}
              </el-radio>
            </div>
          </el-col>
        </el-row>

        <h3 style="display: inline-block;">
          版本
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(value,key) in vipList" style="margin:5px;">
            <div>
              <el-radio size="medium " v-model="chooseVipLevel" @change="queryVipInfo" :label="value.level">
                {{ value.name }}
              </el-radio>
            </div>
          </el-col>
        </el-row>
        <h3 style="display: inline-block;">
          选择账号
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(value,key) in userNameList" style="margin:5px;">
            <div>
              <el-radio size="medium " v-model="account" @change="queryVipInfo" :label="value.UserName">
                {{ value.Name }}
              </el-radio>
            </div>
          </el-col>
        </el-row>
        <h3 style="display: inline-block;">
          输入天数(支持负值)
        </h3><br/>
        <el-input style="width:300px;margin-right:20px;" v-model="expiredDay" placeholder="天数"></el-input>
        <br/>
        <el-button type="primary" @click="exec()" style="margin-top: 10px;">变更</el-button>

      </div>

    </el-card>
    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult"
              rows="25"></el-input>
    <br/>

  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import redisList from "../config/redisList.json";

let vipList = require("../config/vipList.json")
let systemTypeList = require("../config/systemTypeList.json")
let userNameList = require("../config/userName.json")
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
      chooseVipLevel: 1,
      chooseSystemType: 1,
      //过期时间
      expiredTime: '',
      chooseVipUserName: '',
      //账号信息
      account: '',
      //系统类型
      systemTypeList: systemTypeList,
      //账号
      userNameList: userNameList,
      //vip版本
      vipList: vipList,
      //过期时间
      expiredDay: 10,
      //选择的系统
      chooseUserName: "common3",
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
    //增加uniKey
    for (let i in redisList) {
      redisList[i].UniKey = redisList[i].Name
    }
    this.redisConfigList = redisList
  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        ExecType: 'change_vip_type',
        xkfDevDbConfig: this.xkfDevDbConfig,
        Account: this.account,
        VipLevel: this.chooseVipLevel,
        SystemType: this.chooseSystemType,
        redisConfigList: _that.redisConfigList,
        expiredDay: this.expiredDay,
      }
      if (this.account === '-1' || this.account === '') {
        this.error('请输入或选择账号');
        return
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    queryVipInfo: function () {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        ExecType: 'query_vip_info',
        xkfDevDbConfig: this.xkfDevDbConfig,
        Account: this.account,
        VipLevel: this.chooseVipLevel,
        SystemType: this.chooseSystemType,
        redisConfigList: _that.redisConfigList,
        expiredDay: this.expiredDay,
      }
      if (this.account === '-1' || this.account === '' || this.chooseSystemType === '' || this.chooseSystemType === -1) {
        return
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
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
