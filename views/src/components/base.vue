<template>

</template>

<script>
export default {
  data () {
    return {

    }
  },
  methods : {
    setLoading : function (params){
      if(!this.$refs[params.ExecType]){
        return
      }
      this.$refs[params.ExecType].loading = true
    },
    cancelLoading : function (params){
      if(!this.$refs[params.ExecType]){
        return
      }
      this.$refs[params.ExecType].loading = false
    },
  },
}
</script>
