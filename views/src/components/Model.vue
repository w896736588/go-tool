<template>
  <el-card>
    <el-card>
    <div>
      <h3 style="display: inline-block;">
        Model生成(下方贴入navicat建表语句，见示例)
      </h3>
      <el-button type="primary" @click="exec()">生成</el-button>
    </div>
    <br/>

    <el-input style="margin-top: 20px;" id="resultTextarea" :placeholder="modelSqlPlaceholder" type="textarea" v-model="modelSql" rows="10"></el-input>
  </el-card>
    <el-input style="margin-top: 20px;" id="modelTextarea" type="textarea" v-model="modelResult" rows="20"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

export default {
  data() {
    return {
      name: "Model",
      modelSql: "",
      modelSqlPlaceholder: "CREATE TABLE `tbl_kf_custom_auto_receive_num` (\n" +
        "  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n" +
        "  `admin_user_id` int(10) unsigned NOT NULL DEFAULT '0',\n" +
        "  `wechatapp_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '应用ID，如果为0 ，那么表示通用设置',\n" +
        "  `receive_num` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '设置的接待人数',\n" +
        "  `is_switch` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '是否开启，1：开启，0：关闭',\n" +
        "  `create_time` int(10) unsigned NOT NULL DEFAULT '0',\n" +
        "  `update_time` int(10) unsigned NOT NULL DEFAULT '0',\n" +
        "  PRIMARY KEY (`id`)\n" +
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自定义自动接待人数';",
      modelResult: "",
      sshConfig: {},
      templateModel: "/**\n" +
        " * {table_desc}\n" +
        " * @author {author}\n" +
        " * @date {date}\n" +
        " */\n" +
        "class {table_class} extends BaseModel {\n" +
        "\n" +
        "    public function __construct($db = null) {\n" +
        "        parent::__construct($db);\n" +
        "        $this->table = '{table_name}';\n" +
        "        $this->cols  = [\n" +
        "{cols}\n" +
        "        ];\n" +
        "    }\n" +
        "}",
    }
  },
  mounted: function () {
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig !== null) {
      this.sshConfig = JSON.parse(sshConfig)
    }
  },
  methods: {
    //执行
    exec: function () {
      if(this.modelSql === ""){
        this.error('请输入建表语句')
        return
      }
      let tableName = this.getTableName()
      let tableClassName = this.getTableClassName(tableName)
      let tableDesc = this.getTableDesc(tableName)
      let author = this.sshConfig.username
      let date = this.getCurrentDate()
      let cols = this.getCols()

      let modelResult = this.templateModel
      modelResult = modelResult.replaceAll("{table_desc}" , tableDesc)
      modelResult = modelResult.replaceAll("{table_class}" , tableClassName)
      modelResult = modelResult.replaceAll("{author}" , author)
      modelResult = modelResult.replaceAll("{date}" , date)
      modelResult = modelResult.replaceAll("{table_name}" , tableName)
      modelResult = modelResult.replaceAll("{cols}" , cols)
      this.modelResult = "<?php \n" + modelResult
    },
    getTableName: function () {
      let reg = /CREATE TABLE `[a-zA-z_0-9]+` [(]/;
      let matchResult = this.modelSql.match(reg);
      let tableName = ""
      if (matchResult.length > 0) {
        tableName = matchResult[0]
        tableName = tableName.replaceAll("CREATE TABLE `", "")
        tableName = tableName.replaceAll("` (", "")
      }
      return tableName
    },
    getCols : function (){
      let colsList = this.modelSql.split("\n")
      let regField = /^  `[a-zA-z_0-9]+`/;
      let regFieldComment = /COMMENT '.+',/
      colsList.pop()
      colsList.shift()
      let returnColList = []
      for(let i in colsList){
        let col = colsList[i]
        let fieldNameResult = col.match(regField)
        let fieldName = ""
        if (fieldNameResult && fieldNameResult.length > 0) {
          fieldName = fieldNameResult[0]
          fieldName = fieldName.replaceAll("`", "")
        }
        fieldName = fieldName.replaceAll(" " , "")
        if(fieldName === ""){
          continue
        }

        let fieldNameCommentResult = col.match(regFieldComment)
        let fieldNameComment = ""
        if (fieldNameCommentResult && fieldNameCommentResult.length > 0) {
          fieldNameComment = fieldNameCommentResult[0]
          fieldNameComment = fieldNameComment.replaceAll("COMMENT '", "")
          fieldNameComment = fieldNameComment.replaceAll("',", "")
        }
        if(fieldNameComment === ""){
          fieldNameComment = fieldName
        }
        let prefix = "           '" + fieldName + "',"
        //计算空格长度
        let space = ""
        for(let j = 0 ; j < 50 - prefix.length ; j++ ){
          space += " "
        }
        returnColList.push(prefix + space + "//" + fieldNameComment)
      }
      console.log(returnColList)
      return returnColList.join("\n")
    },
    getCurrentDate: function () {
      let date = new Date()
      let year = date.getFullYear() //获取完整的年份(4位)//
      let month = date.getMonth() + 1 //获取当前月份(0-11,0代表1月)
      let strDate = date.getDate() // 获取当前日(1-31)
      if (month < 10) {
        month = `0${month}` // 如果月份是个位数，在前面补0
      }
      if (strDate < 10) {
        strDate = `0${strDate}` // 如果日是个位数，在前面补0
      }
      return `${year}-${month}-${strDate}`
    },
    getTableDesc: function (tableName) {
      let reg = /COMMENT='.+';/;
      let matchResult = this.modelSql.match(reg);
      if(matchResult === null){
        return tableName
      }
      let tableDesc = ""
      if (matchResult.length > 0) {
        tableDesc = matchResult[0]
        tableDesc = tableDesc.replaceAll("COMMENT='", "")
        tableDesc = tableDesc.replaceAll("';", "")
      }
      console.log(tableDesc)
      return tableDesc
    },
    getTableClassName: function (tableName) {
      let tableClassName = tableName.replaceAll("tbl_", "")
      let tableClassList = tableClassName.split("_")
      let returnTableClassList = []
      for (let i in tableClassList) {
        returnTableClassList.push(tableClassList[i].slice(0, 1).toUpperCase() + tableClassList[i].slice(1).toLowerCase())
      }
      console.log(returnTableClassList.join("") + "Model")
      return returnTableClassList.join("") + "Model"
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
