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
        <el-button type="primary" :loading="loadingStatus['change_vip_type']" @click="exec()" style="margin-top: 10px;">变更</el-button>
        <el-button type="primary" :loading="loadingStatus['query_vip_info']" @click="queryVipInfo()" style="margin-top: 10px;">查询</el-button>

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

let vipList = require("../config/zhima/vipList.json")
let systemTypeList = require("../config/zhima/systemTypeList.json")
let userNameList = require("../config/zhima/userName.json")
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
      loadingStatus : {},
    }
  },
  mounted: function () {
    this.apiHost = this.$helperConfig.getApiHost()
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.xkfDevDbConfig = this.$helperConfig.getXkfDevDbConfig()
    this.redisConfigList = this.$helperConfig.getRedisList()
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
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
        _that.$helperNotify.error('请输入或选择账号');
        return
      }
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
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
      this.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
        _that.execResult = response.Data
      });
    },
    setLoading: function (params) {
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function () {
        that.loadingStatus[params.ExecType] = false
      }, 25000)
    },
    cancelLoading: function (params) {
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 1000)
    },
  },
}
</script>

<style scoped>

</style>
