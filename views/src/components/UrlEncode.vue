<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "UrlEncode",
      src : "http://www.jsons.cn/urlencode/",
    }
  },
}
</script>

<style scoped>

</style>
