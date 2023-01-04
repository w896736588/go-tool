<template>
  <el-card>
    <el-select v-model="chooseParentType" @change="changeParentType" placeholder="请选择系统">
      <el-option
        v-for="(value,key) in parentTypeList"
        :key="value.Name"
        :label="value.Title"
        :value="value.Name">
      </el-option>
    </el-select>

    <!--    分支名-->
    <el-input style="width:300px;margin-right:20px;"
              v-model="BranchName" placeholder="请输入分支名"></el-input>

    <el-button type="primary" :loading="btnLoading.change" @click="gitOpType = 'change_branch';exec()" >切换分支</el-button>



    <!--      代码环境-->
    <div style="margin-top: 10px;">
      <h3>{{chooseParentName}} - {{chooseEvnName}} &nbsp;<el-button type="primary" size="mini" :loading="btnLoading.pull" @click="gitOpType = 'pull_branch_origin';exec()">↓ {{chooseEvnName}} pull</el-button></h3>
      <el-row :gutter="20">
        <el-col :span="4" v-for="(value,key) in codeEnvList" style="margin:5px;" v-if="value.ParentType === chooseParentType">
          <div>
            <el-radio @change="codeChange" size="medium " v-model="chooseEvnName" :label="value.Name">{{value.Name}}</el-radio>
          </div>
        </el-col>
      </el-row>
    </div>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
let codeList = require("../config/codeList.json")
let dockerList = require("../config/dockerList.json")
let businessTypeList = require("../config/businessTypeList.json")
export default {
  data() {
    return {
      name: "Consumer",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {},
      prodTestSshConfig : {},
      //选中的环境
      chooseEvnName: "common3",
      //代码环境
      codeEnvList: codeList,
      //docker
      dockerList: dockerList,
      //按钮状态
      btnLoading : {
        exec : false,
        pull : false,
        change : false,
      },
      //操作业务类型
      chooseBusinessType: "git",
      businessTypeList: businessTypeList,
      //操作父类型
      chooseParentType: "xkf",
      chooseParentName : "小客服（php）",
      parentTypeList: [
        {Title: "小客服（php）", Name: "xkf"},
        {Title: "企微（php）", Name: "wk"},
        {Title: "视频号小店（golang）", Name: "weixin_shop_golang"},
        // {Title: "预发布", Name: "prodTest"},
      ],
      //总的操作类型
      ExecType: "pull_branch_origin",
      //操作类型
      dialogSshConfig: false,
      BranchName: "",  //分支名
      execResult: "",//操作结果
      gitOpType : 'pull_branch_origin',
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
    let prodTestSshConfig = this.getStore('prodTestSshConfig')
    if (prodTestSshConfig !== null) {
      this.prodTestSshConfig = JSON.parse(prodTestSshConfig)
    }
  },
  methods: {
    //改变代码环境
    codeChange : function (){
      console.log(this.chooseEvnName)
      this.gitOpType = 'query_current_branch'
      this.exec()
    },
    //改变git操作类型
    gitOpTypeChange : function (){
      console.log(this.ExecType)
      this.ExecType = this.gitOpType
    },
    //改变父类类型
    changeParentType: function () {
      this.chooseEvnName = ''
      this.ExecType = ''
      this.chooseBusinessType = ''
      for(let i in this.parentTypeList){
        if(this.parentTypeList[i].Name === this.chooseParentType){
          this.chooseParentName = this.parentTypeList[i].Title
        }
      }
    },
    //改变业务类型
    changeBusinessType: function () {
      this.chooseEvnName = ""
      this.ExecType = ''
    },
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = {};
      for (let i in this.codeEnvList) {
        if (this.codeEnvList[i].Name === this.chooseEvnName && this.codeEnvList[i].ParentType === this.chooseParentType) {
          env_config = this.codeEnvList[i]
          break
        }
      }
      if (env_config === {} || env_config === undefined || env_config === 'undefined' || !env_config) {
        _that.error("不存在的配置");
        return
      }
      if(!_that.sshConfig || !_that.sshConfig.username || _that.sshConfig.username === ''){
        _that.error("请先配置ssh");
        return
      }
      if(this.CodePath === ''){
        _that.error("请选择代码环境");
        return
      }
      this.ExecType = this.gitOpType
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        CodePath: env_config.CodePath,
        BranchName: this.BranchName,
        ExecType: this.ExecType,
        WechatKefuAppid: this.chooseWechatKefuAppid,
        DockerList: this.dockerList,
        ParentType : env_config.ParentType,
        DockerId: "",
        DockerCodePath: env_config.DockerCodePath,
      }
      if (params.ExecType === 'change_branch' && params.BranchName === '') {
        _that.error('分支名不能为空')
        return
      } else if (params.ExecType === 'supervisor_restart_all' && params.CodePath === '') {
        _that.error('请选择代码环境')
        return
      }

      //如果是切换微信客服 需要找到code对应的docker
      for (let j in this.dockerList) {
        if (env_config.DockerName === this.dockerList[j].Name) {
          params.DockerId = this.dockerList[j].Id
        }
      }
    if (params.ExecType === 'supervisor_restart_all' && params.DockerId === ``) {
        _that.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
      });
    },
    setBtnLoading : function (params){
      if(params.ExecType === 'pull_branch_origin'){
        this.btnLoading.pull = true
      }else if (params.ExecType === 'change_branch'){
        this.btnLoading.change = true
      }

      let _this = this
      let _set_params = params
      setTimeout(function (){
        _this.cancelBtnLoading(_set_params)
      } , 15000)
    },
    cancelBtnLoading : function (params){
      if(params.ExecType === 'pull_branch_origin'){
        this.btnLoading.pull = false
      }else if (params.ExecType === 'change_branch'){
        this.btnLoading.change = false
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
