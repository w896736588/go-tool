<template>
  <el-card>

    <el-card class="box-card">
      <h3 style="display: inline-block;">
        客服系统登录
      </h3>
      <div v-for="(value,k) in userNameList" :key="k" class="text item" style="margin-top:15px;">
        {{value.Name}}
          <el-link type="primary" v-for="(valueBtn,key) in loginTypeList" @click="login(value,valueBtn)" style="margin-left:10px;">
            {{valueBtn.loginName}}
          </el-link>
      </div>
    </el-card>

    <el-card class="box-card" v-for="(valueLink,k) in linkList" style="margin-top: 10px;">
      <h3 style="display: inline-block;">
        {{valueLink.name}}
      </h3>
      <el-row :gutter="20">
        <el-col v-for="(valueLinkValue,k) in valueLink.list" :key="k" :span="3">
          <div class="grid-content bg-purple">
            <el-link type="primary" @click="redirectLink(valueLinkValue)" style="margin-left:10px;">
              {{valueLinkValue.name}}
            </el-link>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import redisList from "../config/redisList.json";
let vipList = require("../config/vipList.json")
let systemTypeList = require("../config/systemTypeList.json")
let userNameList = require("../config/userName.json")
let linkList = require("../config/urlList.json")
export default {
  data() {
    return {
      name: "Vip",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      xkfDevDbConfig : {},
      sshConfig: {},
      prodTestSshConfig : {},
      //选中的vip版本
      chooseVipLevel : -1,
      chooseSystemType : -1,
      //过期时间
      expiredTime : '',
      chooseVipUserName : '',
      //账号信息
      account : '',
      //系统类型
      systemTypeList : systemTypeList,
      //账号
      userNameList : userNameList,
      //vip版本
      vipList : vipList,
      //过期时间
      expiredDay : 10,
      //选择的系统
      chooseUserName : "common3",
      chooseLoginType : "1",
      loginTypeList : [
        {
          "loginName" : "主环境",
          "value" : "1",
        },
        {
          "loginName" : "子环境",
          "value" : "2",
        },
        {
          "loginName" : "主环境运营后台",
          "value" : "3",
        },
        {
          "loginName" : "子环境运营后台",
          "value" : "4",
        }
      ],
      //总的操作类型
      ExecType: "",
      execResult: "",//操作结果
      redisConfigList : [],
      linkList : linkList
    }
  },
  mounted: function () {
    if(process.env.NODE_ENV === 'production'){
      this.apiHost = '';
    }
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig !== null) {
      this.sshConfig = JSON.parse(sshConfig)
    }
    if(!this.sshConfig || !this.sshConfig.username || this.sshConfig.username === ''){
      this.error("请先配置ssh");
      return
    }
    let xkfDevDbConfig = this.getStore('devTestDbConfig')
    if (xkfDevDbConfig !== null) {
      this.xkfDevDbConfig = JSON.parse(xkfDevDbConfig)
    }
    //增加uniKey
    for( let i in redisList){
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
        xkfDevDbConfig : this.xkfDevDbConfig,
        Account : this.account,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList,
        expiredDay : this.expiredDay,
      }
      if(this.account === '-1' || this.account === ''){
        this.error('请输入或选择账号');
        return
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    redirectLink : function (linkValue){
      this.execResult = linkValue.link
      window.open(this.execResult,'_blank');
    },
    //登录
    login : function (userValue,loginTypeValue){
      let _that = this
      let loginUrl = ''
      if(loginTypeValue.value === '1' || loginTypeValue.value === '2'){
        loginUrl = '/index/index';
      }else{
        loginUrl = '/XkfOperate/CustomerList';
      }
      let loginHost = ``
      if(loginTypeValue.value === '1' || loginTypeValue.value === '3'){
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === userValue.Name){
            loginHost = this.userNameList[i].Host
          }
        }
      }else{
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === userValue.Name){
            loginHost = this.userNameList[i].HostChild
          }
        }
      }
      let account = ''
      for(let i in this.userNameList){
        if(this.userNameList[i].Name === userValue.Name){
          account = this.userNameList[i].UserName
        }
      }
      let params = {
        Account : account,
        loginUrl : loginUrl,
        loginHost : loginHost,
        SshConfig: _that.sshConfig,
        ExecType: 'login_xkf',
        xkfDevDbConfig : this.xkfDevDbConfig,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList,
      }
      if(loginTypeValue.value === '3' || loginTypeValue.value === '4'){
        params.Account = '2@163.com'
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        window.open(response.Data,'_blank');
      });
    },
    queryVipInfo : function (){
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        ExecType: 'query_vip_info',
        xkfDevDbConfig : this.xkfDevDbConfig,
        Account : this.account,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList,
        expiredDay : this.expiredDay,
      }
      if(this.account === '-1' || this.account === '' || this.chooseSystemType === '' || this.chooseSystemType === -1){
        return
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    success: function (msg) {
      // Message.success(msg);
      this.$notify({title: '提示', message: msg, type: 'success' , duration : 1000});
    },
    warning: function (msg) {
      // Message.warning(msg);
      this.$notify({title: '提示', message: msg, type: 'warning' , duration : 1000});
    },
    info: function (msg) {
      // Message.info(msg);
      //this.$notify({title: '提示', message: msg});
      this.$notify({title: '提示', message: msg, type: 'info' , duration : 1000});
    },
    error: function (msg) {
      // Message.error(msg);
      this.$notify({title: '提示', message: msg, type: 'error' , duration : 1000});
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
