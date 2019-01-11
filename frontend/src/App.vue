<template>
  <div id="app">
    <!-- <img alt="Vue logo" src="./assets/logo.png"> -->
    <el-container>
      <el-header>mp3 -> m4r -> download</el-header>
      <el-main>
        <Audio/>
      </el-main>
      <el-footer>
        <!-- Footer -->
      </el-footer>
    </el-container>
  </div>
</template>

<script>
import Audio from "./components/Audio.vue";

export default {
  name: "app",
  components: {
    Audio
  }
};
</script>

<style>
#app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
