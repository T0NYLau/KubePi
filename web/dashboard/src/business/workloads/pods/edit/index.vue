<template>
  <div v-loading="loading">
    <yaml-editor :value="item" :is-edit="true" ref="yaml_editor"></yaml-editor>
    <div class="bottom-button">
      <el-button @click="$emit('close')">{{ $t("commons.button.cancel") }}</el-button>
      <el-button v-loading="loading" @click="onSubmit" type="primary">
        {{ $t("commons.button.submit") }}
      </el-button>
    </div>
  </div>
</template>

<script>
import YamlEditor from "@/components/yaml-editor"
import {getPodByName, updatePod} from "@/api/pods"

export default {
  name: "PodEdit",
  components: { YamlEditor },
  props: {
    name: String,
    namespace: String
  },
  data () {
    return {
      loading: false,
      item: {},
    }
  },
  methods: {
    getDetail () {
      this.loading = true
      const cluster = this.$route.query.cluster
      getPodByName(cluster, this.namespace, this.name).then(res => {
        this.item = res
      }).finally(() => {
        this.loading = false
      })
    },
    onSubmit () {
      this.loading = true
      const data = this.$refs.yaml_editor.getValue()
      const cluster = this.$route.query.cluster
      updatePod(cluster, this.namespace, this.name, data).then(() => {
        this.$emit('success')
      }).finally(() => {
        this.loading = false
      })
    }
  },
  created () {
    this.getDetail()
  }
}
</script>

<style scoped>

</style>
