<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Ip",
      src : "https://ip.tool.chinaz.com/github.com",
    }
  },
}
</script>

<style scoped>

</style>
