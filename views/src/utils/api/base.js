//登录
import Vue from "vue";
import store from "../store";
import module from "../api/module"

//登录拿到 unikey
function BaseLogin(userName, password) {
  BasePost('/api/BaseLogin', {
    UserName: userName,
    Password: password,
  }, function (response) {
    console.log(response)
    store.setStore('Unikey', response.Data.unikey)
    BaseRegisterService()
  })
}

//注册服务
function BaseRegisterService() {
  let redisConfigList = module.GetRedisConfigList()
  let mysqlConfigList = module.GetMysqlConfigList()
  let shellConfigList = module.GetShellConfigList()
  let encryptConfig = module.GetEncryptConfig()
  BasePost('/api/BaseRegisterService', {
    Unikey: store.getStore('Unikey'),
    MysqlConfigList : mysqlConfigList,
    RedisConfigList : redisConfigList,
    ShellConfigList : shellConfigList,
    EncryptKey : encryptConfig['EncryptKey'],
    EncryptIv : encryptConfig['EncryptIv']
  }, function (response) {


  })
}

//检查unikey是否已经登录注册
function BaseCheckService() {
  console.log('baseCheckService')
  let unikey = store.getStore('Unikey')
  if(unikey === undefined || unikey === null || unikey === 'null' || unikey === 'undefined'){
    let userName = store.getStore('UserName')
    let password = store.getStore('Password')
    userName = '1';
    password = '1';
    BaseLogin(userName, password)
    return
  }
  BasePost('/api/BaseCheckUnikeyExist', {
    Unikey: unikey,
  }, function (response) {
    if (response.Data.NeedLogin === '1') {
      let userName = store.getStore('UserName')
      let password = store.getStore('Password')
      BaseLogin(userName, password)
    }
  })
}

//POST请求
function BasePost(uri, params, callBack) {
  Vue.axios.post(GetApiHost() + uri, params).then(function (response) {
    callBack(response)
  });
}

//拿到接口地址
function GetApiHost() {
  if (process.env.NODE_ENV === 'production') {
    return '';
  }
  return 'http://localhost:7073';
}

export default {
  BaseLogin,
  BaseRegisterService,
  BasePost,
  BaseCheckService,
}
