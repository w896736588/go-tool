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

<!--    业务类型-->
    <el-select v-model="chooseBusinessType" @change="changeBusinessType" placeholder="请选择操作业务">
      <el-option
        v-for="(value,key) in businessTypeList"
        :key="value.Name"
        :label="value.Title"
        :value="value.Name">
      </el-option>
    </el-select>

    <!--    微信客服-->
    <el-select v-model="chooseWechatKefuAppid" placeholder="请选择微信客服" v-if="chooseBusinessType === 'wechat_kefu' && chooseParentType === 'xkf'">
      <el-option
        v-for="(value,key) in wechatKefuList"
        :key="value.name"
        :label="value.name"
        :value="value.appid">
      </el-option>
    </el-select>

<!--    环境-->
    <el-select v-model="chooseEvnName" placeholder="请选择代码环境">
      <el-option
        v-for="(value,key) in env.codeEnvList" v-if="value.ParentType === chooseParentType"
        :key="value.Name"
        :label="value.Name"
        :value="value.Name">
      </el-option>
    </el-select>

<!--    git操作类型-->
    <el-select v-model="ExecType" v-if="chooseBusinessType === 'git'" placeholder="请选择git操作">
      <el-option
        v-for="(value,key) in gitOpTypeList"
        :key="value.ExecType"
        :label="value.Name"
        :value="value.ExecType">
      </el-option>
    </el-select>



<!--    分支名-->
    <el-input v-if="ExecType === 'change_branch' && chooseBusinessType === 'git'" style="width:300px;margin-right:20px;" v-model="BranchName"
              placeholder="请输入分支名"></el-input>

    <el-button type="primary" @click="exec" v-if="chooseBusinessType === 'git'">执 行</el-button>
    <el-button type="primary" icon="el-icon-setting" @click="dialogSshConfig = !dialogSshConfig">SSH设置</el-button>
    <el-card style="margin-top: 20px;" v-if="dialogSshConfig">
      <el-form ref="form" :model="sshConfig" label-width="80px">
        <el-form-item label="账号名">
          <el-input v-model="sshConfig.username"></el-input>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="sshConfig.password"></el-input>
        </el-form-item>
        <el-form-item label="IP地值">
          <el-input v-model="sshConfig.host"></el-input>
        </el-form-item>
        <el-form-item label="端口">
          <el-input v-model="sshConfig.port"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSshConfig">保存配置</el-button>
          <el-button @click="dialogSshConfig = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>

<!--  子操作选项列表-->
    <el-card style="margin-top: 20px;" v-if="chooseBusinessType === 'wechat_kefu'">
      微信客服操作列表<br/><br/>
      <el-button type="primary" @click="ExecType = 'wechat_kefu_change';exec()">切换到当前选择的环境</el-button>
      <el-button type="primary" @click="ExecType = 'wechat_kefu_status';exec()">查看所在环境</el-button>
<!--      <el-button type="primary" @click="">查看错误日志</el-button>-->
    </el-card>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>
  <!--    <el-card>-->
  <!--      <el-input type="textarea" v-model="execResult" rows="20"></el-input>-->
  <!--    </el-card>-->
</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

export default {
  data() {
    return {
      name: "Consumer",
      tabPosition: 'left',
      ws: null,
      heartInternal: null,
      wsUrl: 'ws://localhost:7778/redisWebSocket',
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {
        username: "frog",
        password: "frog987^%$321_220",
        host: "121.40.109.241",
        port: "22",
      },
      //选中的环境
      chooseEvnName: "common3",
      //按环境
      env: {
        codeEnvList: [
          {Name: "common3",   ParentType : "xkf", DockerName : "common3", DockerCodePath : "apps/yii_customer_service",  CodePath: "docker_apps/common3/yii_customer_service"},
          {Name: "common31",  ParentType : "xkf", DockerName : "common3", DockerCodePath : "apps/yii_customer_service_sub01", CodePath: "docker_apps/common3/yii_customer_service_sub01"},
          {Name: "common1",   ParentType : "xkf", DockerName : "common1", DockerCodePath : "apps/yii_customer_service", CodePath: "docker_apps/common1/yii_customer_service"},
          {Name: "common11",  ParentType : "xkf", DockerName : "common1", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common1/yii_customer_service_sub01"},
          {Name: "common",    ParentType : "xkf", DockerName : "common", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common/yii_customer_service"},
          {Name: "common01",  ParentType : "xkf", DockerName : "common", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common/yii_customer_service_sub01"},
          {Name: "common2",   ParentType : "xkf", DockerName : "common2", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common2/yii_customer_service"},
          {Name: "common21",  ParentType : "xkf", DockerName : "common2", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common2/yii_customer_service_sub01"},
          {Name: "common4",   ParentType : "xkf", DockerName : "common4", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common4/yii_customer_service"},
          {Name: "common41",  ParentType : "xkf", DockerName : "common4", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common4/yii_customer_service_sub01"},
          {Name: "common5",   ParentType : "xkf", DockerName : "common5", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common5/yii_customer_service"},
          {Name: "common51",  ParentType : "xkf", DockerName : "common5", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common5/yii_customer_service_sub01"},
          {Name: "common6",   ParentType : "xkf", DockerName : "common6", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common6/yii_customer_service"},
          {Name: "common61",  ParentType : "xkf", DockerName : "common6", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common6/yii_customer_service_sub01"},
          {Name: "common7",   ParentType : "xkf", DockerName : "common7", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/common7/yii_customer_service"},
          {Name: "common71",  ParentType : "xkf", DockerName : "common7", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/common7/yii_customer_service_sub01"},
          {Name: "mike",   ParentType : "xkf", DockerName : "mike", DockerCodePath : "apps/yii_customer_service",CodePath: "docker_apps/mike/yii_customer_service"},
          {Name: "mike1",  ParentType : "xkf", DockerName : "mike", DockerCodePath : "apps/yii_customer_service_sub01",CodePath: "docker_apps/mike/yii_customer_service_sub01"},
        ],
      },
      //docker
      dockerList : [
        {Name : "common3", Id : "18ecd50ae1fa"},
        {Name : "common1", Id : "d8272fe87a7c"},
        {Name : "common5", Id : "a5baf7b8a7d6"},
        {Name : "common", Id : "80e8a36d2fda"},
        {Name : "common6", Id : "7356a3e48e97"},
        {Name : "common7", Id : "75ac41ec28f8"},
        {Name : "common4", Id : "3ee4a2fe6420"},
        {Name : "common2", Id : "43c29634dd7f"},
        {Name : "mike", Id : "a1de0e28a81a"},
      ],
      //操作业务类型
      chooseBusinessType : "git",
      businessTypeList : [
        {Title : "git操作", Name : "git"},
        {Title : "微信客服操作", Name : "wechat_kefu"},
        {Title : "消费者管理", Name : "supervisor"},
        {Title : "调整VIP版本", Name : "vip"},
        {Title : "docker管理", Name : "docker"},
      ],
      //操作父类型
      chooseParentType : "xkf",
      parentTypeList : [
        {Title : "小客服", Name : "xkf"},
        {Title : "企微", Name : "wk"},
      ],
      //微信客服合集
      chooseWechatKefuAppid : "",
      wechatKefuList : [
        {appid : "wpX2IKEAAA9PH4WJVe2nQgEfOh7MXD-A", name : "芝麻微客v2(授权接入)"},
        {appid : "wpX2IKEAAA051WWUOg1vUqYoAVZ7PZ_A", name : "武汉芝麻小客服网络(授权接入)"},
        {appid : "wpX2IKEAAABioq2s-opO6ttmo6XGlOBQ", name : "武汉气形网络(授权接入)"},
        {appid : "wpX2IKEAAAfCrs43krHmtfHzxfgbx2lg", name : "零一网络科技（武汉）有限公司客服(授权接入)"},
        {appid : "wpX2IKEAAAgKbVQf07vK27MIku8iRhBw", name : "芝麻小事网络科技(武汉)有限公司客服(授权接入)"},
        {appid : "wpX2IKEAAAwS7tM_udiV9JL4FVibhXpw", name : "上海芝麻小事网络科技有限公司(授权接入)"},
        {appid : "wpX2IKEAAAC90zLn5lOxRfXndUPkf43g", name : "微信客服芝麻小事(授权接入)"},
        {appid : "ww5f432b3a24a9b9f1", name : "武汉芝麻小客服网络客服(密码接入)"},
        {appid : "ww6b983da349fd945e", name : "上海芝麻小事网络科技有限公司(密码接入)"},
        {appid : "ww6f5a28a32b2a0fe9", name : "武汉铁杵磨针网络科技有限公司(授权接入)"},
        {appid : "ww10b0159168693b3f", name : "helen测试企业(密码接入)"},
        {appid : "wwbbaa6ccaf7ef62fb", name : "KKK测测(密码接入)"},
        {appid : "wpX2IKEAAASrpL2F1R3zB3QMPFT0esEw", name : "武汉芝麻小客服(授权接入)"},
      ],
      //总的操作类型
      ExecType: "query_current_branch",
      //操作类型
      dialogSshConfig : false,
      BranchName: "",  //分支名
      execResult: "",//操作结果
      gitOpTypeList: [
        {
          "Name": "查询当前分支",
          "ExecType": "query_current_branch",
        },
        {
          "Name": "更新当前分支到最新代码",
          "ExecType": "pull_branch_origin",
        },
        {
          "Name": "切换分支并更新到最新",
          "ExecType": "change_branch",
        },
      ],
    }
  },
  mounted: function () {
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig === null) {
      this.sshConfig = {
        username: "",
        password: "",
        host: "121.40.109.241",
        port: "22",
      }
    }
  },
  methods: {
    //改变父类类型
    changeParentType : function (){
      this.chooseEvnName = ''
      this.ExecType = ''
      this.chooseBusinessType = ''
    },
    //改变业务类型
    changeBusinessType : function (){
      this.chooseEvnName = ""
      this.ExecType = ''
    },
    //保存ssh配置
    saveSshConfig: function () {
      if(this.sshConfig.username === ''){
        this.error('用户名不能为空')
        return;
      }else if(this.sshConfig.password === ''){
        this.error('密码不能为空')
        return;
      }else if(this.sshConfig.host === ''){
        this.error('连接IP不能为空')
        return;
      }else if(this.sshConfig.port === ''){
        this.error('端口不能为空')
        return;
      }
      this.setStore('sshConfig', JSON.stringify(this.sshConfig))
      this.success('设置成功')
      this.dialogSshConfig = false
    },
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = {};
      for (let i in this.env.codeEnvList) {
        if (this.env.codeEnvList[i].Name === this.chooseEvnName) {
          env_config = this.env.codeEnvList[i]
          break
        }
      }
      if (env_config === {}) {
        _that.error("不存在的配置");
        return
      }
      env_config.SshConfig = _that.sshConfig
      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        BranchName : this.BranchName,
        ExecType : this.ExecType,
        WechatKefuAppid : this.chooseWechatKefuAppid,
        DockerList : this.dockerList,
        DockerId : "",
        DockerCodePath : env_config.DockerCodePath,
      }
      if(params.ExecType === 'change_branch' && params.BranchName === ''){
        _that.error('分支名不能为空')
        return
      }else if(params.ExecType === 'wechat_kefu_status' && params.WechatKefuAppid === ''){
        _that.error('选择微信客服')
        return
      }else if(params.ExecType === 'wechat_kefu_change' && (params.WechatKefuAppid === '' || params.CodePath === '')){
        _that.error('选择微信客服以及代码环境')
        return
      }

      //如果是切换微信客服 需要找到code对应的docker
      for(let j in this.dockerList){
        if(env_config.DockerName === this.dockerList[j].Name){
          params.DockerId = this.dockerList[j].Id
        }
      }
      if(params.ExecType === 'wechat_kefu_change' && params.DockerId === ``){
        _that.error('代码环境找不到对应的docker')
        return
      }

      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
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
