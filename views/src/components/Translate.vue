<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Translate",
      src : "https://fanyi.youdao.com/",
    }
  },
}
</script>

<style scoped>

</style>
