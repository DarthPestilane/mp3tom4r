<template>
  <iframe :src="url" frameborder="0" :hidden="true"></iframe>
</template>

<script>
export default {
  props: {
    url: {
      type: String,
      default: ""
    }
  }
};
</script>
