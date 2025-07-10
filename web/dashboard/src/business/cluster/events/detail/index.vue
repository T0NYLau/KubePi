<template>
  <layout-content :header="$t('commons.form.detail')" :back-to="{name: 'events'}" v-loading="loading">
   
    <el-row>
      <div >
        <yaml-editor :value="yaml" :read-only="true"></yaml-editor>
        <div class="bottom-button">
          <el-button @click="yamlShow=!yamlShow">{{ $t("commons.button.back_detail") }}</el-button>
        </div>
      </div>
    </el-row>
  </layout-content>
</template>

<script>
import LayoutContent from "@/components/layout/LayoutContent"
import { isJSON } from "@/utils/data"
import { getEventsWithNs } from "@/api/events"
import YamlEditor from "@/components/yaml-editor"

export default {
  name: "EventDetail",
  components: { YamlEditor, LayoutContent },
  props: {
    name: String,
    namespace: String,
    cluster: String
  },
  data() {
    return {
      item: {
        metadata: {},
        spec: {
        },
        status: {},
      },
      yamlShow: false,
      loading: false,
      yaml: {},
    }
  },
  methods: {
    getDetail() {
      this.loading = true
      getEventsWithNs(this.cluster, this.namespace,this.name).then((res) => {
        this.loading = false
        this.item = res
        this.yaml = JSON.parse(JSON.stringify(this.item))
      })
    },
    getContent(value) {
      const { Base64 } = require("js-base64")
      const content = Base64.decode(value)
      return JSON.parse(content)
    },
    jsonV(str) {
      const { Base64 } = require("js-base64")
      const content = Base64.decode(str)
      return isJSON(content)
    },
    getValue(value) {
      const { Base64 } = require("js-base64")
      return Base64.decode(value)
    },
  },
  watch: {
    yamlShow: function (newValue) {
      // 使用window.location.origin获取完整的基础URL
      const baseUrl = window.location.origin
      // 获取当前的公共路径前缀
      const publicPath = process.env.VUE_APP_PUBLIC_PATH || '/'
      
      // 直接构建URL而不是使用router.resolve，确保参数正确传递
      const url = `${baseUrl}${publicPath}events/detail/${this.cluster}/${this.namespace}/${this.name}?yamlShow=${newValue}&cluster=${this.cluster}`
      
      window.open(url, "_blank")
      this.getDetail()
    },
  },
  created() {
    this.getDetail()
  },
}
</script>

<style scoped>
</style>
