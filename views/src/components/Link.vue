<template>
  <el-card>
    <el-card>
    <div>
      <h3 style="display: inline-block;">
        客服系统登录
      </h3>
      <el-row :gutter="20">
        <el-col :span="2" v-for="(value,k) in userNameList" style="margin:5px;">
          <div>
            <el-radio size="medium" @change="changeCodeEnv" v-model="chooseEvnName" :label="value.Name">
              {{ value.Name }}
            </el-radio>
          </div>
        </el-col>
      </el-row>
      <br/>
      <el-input style="width:300px;display: inline-block;margin-top:5px;" v-model="inputAccount" placeholder="输入账号名登录"></el-input>
      <br/>
      <br/>
      <el-button type="primary" :loading="loadingStatus['login_xkf' + value.value]" size="medium" v-for="(value,k) in loginTypeList" @click="login(value)">{{value.loginName}}</el-button>
      <br/>
      <br/>
      <el-button type="primary" :loading="loadingStatus['login_xkf' + value.value]" plain size="medium" v-for="(value,k) in loginTypeChildList" @click="login(value)">{{value.loginName}}</el-button>

<!--      <div v-for="(value,k) in userNameList" :key="k" class="text item" style="margin-top:10px;">-->
<!--        {{value.Name}}-->
<!--        <el-link type="primary" v-for="(valueBtn,key) in loginTypeList" @click="login(value,valueBtn)" style="margin-left:10px;">-->
<!--          {{valueBtn.loginName}}-->
<!--        </el-link>-->
<!--        <el-radio @change="codeChange(value)" size="medium " v-model="chooseEvnName" :label="value.Name">-->
<!--          {{ value.NameTitle }}-->
<!--        </el-radio>-->
<!--      </div>-->


<!--      <div v-for="(value,k) in userNameList" :key="k" class="text item" style="margin-top:10px;">-->
<!--        {{value.Name}}-->
<!--          <el-link type="primary" v-for="(valueBtn,key) in loginTypeList" @click="login(value,valueBtn)" style="margin-left:10px;">-->
<!--            {{valueBtn.loginName}}-->
<!--          </el-link>-->
<!--      </div>-->

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
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

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
      inputAccount : '',
      //账号
      userNameList : [],
      chooseEvnName : "common3",
      chooseLoginType : "1",
      loginTypeList : [
        {
          "loginName" : "主环境后台",
          "value" : "1",
          "url" : "/workbench/adminindex?parent_nav_label=admin_workbench&nav_label=admin_workbench&wechatapp_id={wechatapp_id}&channel_id={channel_id}",
        },
        {
          "loginName" : "主环境运营后台",
          "value" : "3",
          "url" : "/XkfOperate/CustomerList",
        },
        {
          "loginName" : "主环境聊天界面",
          "value" : "5",
          "url" : "/message/chat/#/chat/receive?wechatapp_id={wechatapp_id}",
        },
        {
          "loginName" : "主环境视频号小店",
          "value" : "7",
          "url" : "/accountNumber/workbenches",
        },
      ],
      loginTypeChildList : [
        {
          "loginName" : "子环境后台",
          "value" : "2",
          "url" : "/workbench/adminindex?parent_nav_label=admin_workbench&nav_label=admin_workbench&wechatapp_id={wechatapp_id}&channel_id={channel_id}",
        },
        {
          "loginName" : "子环境运营后台",
          "value" : "4",
          "url" : "/XkfOperate/CustomerList",
        },
        {
          "loginName" : "子环境聊天界面",
          "value" : "6",
          "url" : "/message/chat/#/chat/receive?wechatapp_id={wechatapp_id}",
        },
        {
          "loginName" : "子环境视频号小店",
          "value" : "8",
          "url" : "/accountNumber/workbenches",
        },
      ],
      //总的操作类型
      ExecType: "",
      execResult: "",//操作结果
      linkList : [],
      loadingStatus : {},
    }
  },
  mounted: function () {
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.apiHost = this.$helperConfig.getApiHost()
    this.xkfDevDbConfig = this.$helperConfig.getXkfDevDbConfig()
    this.userNameList = this.$helperConfig.getUsernameList()
    this.linkList = this.$helperConfig.getLinkList()
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
    for(let i in this.loginTypeList){
      this.loadingStatus['login_xkf' + this.loginTypeList[i].value] = false;
    }
    for(let i in this.loginTypeChildList){
      this.loadingStatus['login_xkf' + this.loginTypeChildList[i].value] = false;
    }
  },
  methods: {
    changeCodeEnv : function (){
      this.inputAccount = ""
    },
    redirectLink : function (linkValue){
      this.execResult = linkValue.link
      window.open(this.execResult,'_blank');
    },
    //登录
    login : function (loginTypeValue){
      if(this.chooseEvnName === ``){
        this.$helperNotify.error('先选择环境')
        return
      }
      let _that = this
      let loginUrl = loginTypeValue.url
      let loginHost = ``
      if(loginTypeValue.value === '1' || loginTypeValue.value === '3' || loginTypeValue.value === '5' || loginTypeValue.value === '7'){
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === this.chooseEvnName){
            loginHost = this.userNameList[i].Host
          }
        }
      }else{
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === this.chooseEvnName){
            loginHost = this.userNameList[i].HostChild
          }
        }
      }
      let account = ''
      if(this.inputAccount !== ''){
        account = this.inputAccount
      }else{
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === this.chooseEvnName){
            account = this.userNameList[i].UserName
          }
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
      _that.setLoading(params , loginTypeValue)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelLoading(params , loginTypeValue)
        window.open(response.Data,'_blank');
      });
    },
    setLoading: function (params , value) {
      this.loadingStatus[params.ExecType + value.value] = true
      let that = this
      setTimeout(function () {
        that.loadingStatus[params.ExecType + value.value] = false
      }, 25000)
    },
    cancelLoading: function (params , value) {
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType+ value.value] = false
      } , 1000)
    },
  },
}
</script>

<style scoped>

</style>
