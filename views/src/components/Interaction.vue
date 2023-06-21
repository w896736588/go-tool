<template>
  <el-card>
    <el-dialog :visible.sync="visible" :title="title" width="80%;" :before-close="beforeClose">
      <el-input style="margin-top: 20px;" type="textarea" rows="25"></el-input>
<!--      <el-button type="text" class="button" style="float: right;" >结束</el-button>-->
      <div slot="footer">
<!--        <el-button >取消</el-button>-->
<!--        <el-button type="primary" >结束</el-button>-->
      </div>
    </el-dialog>
  </el-card>

</template>

<script>
export default {
  name : "Interaction",
  props : {
    visible : {
      type : Boolean,
      default : false,
    },
    title : {
      type : String,
      default : "",
    },
    sshConfig : {
      type : Object,
      default : {},
    },
  },
  data (){
    return {
      randKey : '',
    }
  },
  mounted: function () {

  },
  methods: {
    createShell4 : function (){
      this.created()
    },
    beforeClose : function (done){
      this.$emit('before-close' , done) //向上传递
    },
    //建立链接
    created() {
      let url = this.$helperConfig.getWsHost()
      const socket = new WebSocket(url);

      socket.onmessage = (event) => {
        const message = JSON.parse(event.data);

        this.messages.push(message);
      };

      const message = {
        id: 1,
        text: 'Hello, World!',
      };
      socket.onopen = () => {
        socket.send(JSON.stringify(message));
      };
      setInterval(function () {
        socket.send('ping')
      },30000)
    },
  },
}
</script>

<style scoped>

</style>
