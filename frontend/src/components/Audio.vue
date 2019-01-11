<template>
  <el-col style="min-height:60px;" :span="12" :offset="6" v-loading="loading" element-loading-text="正在解析音频">
    <el-upload
      v-show="!soundReady"
      :on-success="handleUploadSuccess"
      :on-exceed="handleUploadExceed"
      :file-list="[]"
      :action="uploadUrl"
    >
      <el-button slot="trigger" size="small" type="primary">上传文件</el-button>
    </el-upload>
    <div v-show="soundReady">
      <el-form ref="form" label-width="80px" label-position="left">
        <el-form-item label="截取">
          <el-col>
            <el-slider
              v-model="range"
              :step="0.01"
              range
              :min="0"
              :max="~~(totalLen/1000)"
              :format-tooltip="formatRange"
              @change="handleCut"
            ></el-slider>
          </el-col>
        </el-form-item>
        <el-form-item label="首尾">
          <el-col :span="11">
            <el-input v-model="range[0]" :min="1"></el-input>
          </el-col>
          <el-col :span="1">~</el-col>
          <el-col :span="11">
            <el-input v-model="range[1]" :min="~~(totalLen/1000)"></el-input>
          </el-col>
          <el-col :span="1">秒</el-col>
        </el-form-item>
        <el-form-item label="时长">
          <el-col style="text-align: left;">{{~~(duration/1000)}} 秒</el-col>
        </el-form-item>
        <el-form-item label="渐入">
          <el-col style="text-align: left;">
            <el-switch v-model="fade"></el-switch>
          </el-col>
        </el-form-item>
      </el-form>
      <el-row :gutter="10">
        <el-button size="small" type="infor" round @click="play">播放</el-button>
        <el-button size="small" type="infor" round @click="pause">暂停</el-button>
        <el-button size="small" type="infor" round @click="replay">重放</el-button>
      </el-row>
      <el-row :gutter="10" style="margin-top:10px;">
        <el-button size="small" type="success" @click="makeM4r" :loading="downloadStart">下载为iPhone铃声</el-button>
        <el-button size="small" type="danger" @click="reUpload">重新上传</el-button>
      </el-row>
    </div>
    <Downloader :url="downloadUrl"/>
  </el-col>
</template>

<script>
import { Howl } from "howler";
import axios from "axios";
import Downloader from "./Downloader.vue";
import config from "../config.js";

export default {
  data() {
    return {
      file: {
        hash: "",
        src: ""
      },
      soundReady: false,
      fade: false,
      startAt: 0, // ms
      duration: 0, // ms
      totalLen: 0, // ms
      range: [0, 0], // second
      howl: new Howl({ src: [""] }),
      downloadStart: false,
      loading: false,
      downloadUrl: "",
      uploadUrl: `${config.ApiUri()}/upload`
    };
  },

  components: {
    Downloader
  },

  methods: {
    initData() {
      this.file = {
        hash: "",
        src: ""
      };
      this.downloadUrl = "";
      this.fade = this.downloadStart = this.loading = this.soundReady = false;
      this.startAt = this.duration = this.totalLen = 0;
      this.range = [0, 0];
    },
    reUpload() {
      this.initData();
    },
    handleUploadSuccess(resp) {
      var data = resp.data;
      this.file.hash = data.hash;
      this.file.src = `${config.AudioUri()}/${data.hash}.mp3`;
      this.initHowl();
    },
    handleUploadExceed() {
      this.$message.warning(`超过限制`);
    },
    initHowl() {
      this.loading = true;
      var howl = new Howl({
        src: [this.file.src],
        onload: () => {
          this.loading = false;
          this.soundReady = true;
          this.duration = howl.duration() * 1000;
          this.totalLen = howl.duration() * 1000;
          this.range = [0, this.duration];
          this.howl = howl;
        }
      });
    },
    play() {
      if (this.howl.playing()) {
        return;
      }
      if (this.fade && this.startAt === this.range[0] * 1000) {
        this.howl.fade(0, 1, 1500);
      }
      var sprite = [this.startAt, this.duration];
      this.howl._sprite = { sprite };
      this.howl.play("sprite");
    },
    pause() {
      if (!this.howl.playing()) {
        return;
      }
      this.startAt = this.howl.pause().seek() * 1000;
      this.howl.stop();
    },
    replay() {
      this.howl.stop();
      this.startAt = this.range[0] * 1000;
      this.play();
    },
    handleCut() {
      this.startAt = this.range[0] * 1000;
      this.duration = (this.range[1] - this.range[0]) * 1000;
    },
    makeM4r() {
      this.downloadStart = true;
      this.pause();
      axios
        .post(`${config.ApiUri()}/convert`, {
          hash: this.file.hash,
          start: this.range[0],
          duration: this.duration / 1000,
          fade: this.fade
        })
        .then(() => {
          this.downloadUrl = `${config.ApiUri()}/download?hash=${
            this.file.hash
          }`;
          this.downloadStart = false;
        });
    },
    formatRange(val) {
      return val + "s";
    }
  }
};
</script>
