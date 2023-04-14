<template>
  <el-card>
    <el-card>
    <div>
      <h3 style="display: inline-block;">
        客服系统登录
      </h3>
      <div v-for="(value,k) in userNameList" :key="k" class="text item" style="margin-top:10px;">
        {{value.Name}}
          <el-link type="primary" v-for="(valueBtn,key) in loginTypeList" @click="login(value,valueBtn)" style="margin-left:10px;">
            {{valueBtn.loginName}}
          </el-link>
      </div>
      <div class="box-card" v-for="(valueLink,k) in linkList" >
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
      </div>
    </div>
  </el-card>
    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import redisList from "../config/redisList.json";

export default {
  data() {
    return {
      name: "Link",
      //接口地址
      apiHost: '',
      //ssh config
      xkfDevDbConfig : {},
      sshConfig: {},
      //账号信息
      account : '',
      //账号
      userNameList : [],
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
      linkList : []
    }
  },
  mounted: function () {
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.apiHost = this.$helperConfig.getApiHost()
    this.xkfDevDbConfig = this.$helperConfig.getXkfDevDbConfig()
    this.userNameList = this.$helperConfig.getUsernameList()
    this.linkList = this.$helperConfig.getLinkList()
  },
  methods: {
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
      }
      if(loginTypeValue.value === '3' || loginTypeValue.value === '4'){
        params.Account = '2@163.com'
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        window.open(response.Data,'_blank');
      });
    },
  },
}
</script>

<style scoped>

</style>
