<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "QrCode",
      src : "https://cli.im/",
    }
  },
}
</script>

<style scoped>

</style>
