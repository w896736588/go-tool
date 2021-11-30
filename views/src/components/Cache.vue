<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Redis",
      src : "/#/redis",
    }
  },
}
</script>

<style scoped>

</style>
