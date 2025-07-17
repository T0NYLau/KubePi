<template>
  <div class="app-container">
    <el-card class="box-card" shadow="never">
      <div slot="header" class="clearfix">
        <span>{{ $t('business.llmmodels.edit_llm_model') }}</span>
      </div>
      <el-form ref="form" :model="form" :rules="rules" label-width="120px" size="small" v-loading="loading">
        <el-form-item :label="$t('business.llmmodels.name')" prop="name">
          <el-input v-model="form.name" disabled />
        </el-form-item>
        <el-form-item :label="$t('business.llmmodels.base_uri')" prop="baseUri">
          <el-input v-model="form.baseUri" placeholder="http://example.com/v1/chat/completions" />
          <div class="el-form-item-tip">请输入完整的API端点URL，例如DeepSeek模型需要包含完整路径: http://ip:port/v1/chat/completions</div>
        </el-form-item>
        <el-form-item :label="$t('business.llmmodels.model_name')" prop="modelName">
          <el-input v-model="form.modelName" />
        </el-form-item>
        <el-form-item :label="$t('business.llmmodels.api_key')" prop="apiKey">
          <el-input v-model="form.apiKey" type="password" show-password />
        </el-form-item>
        <el-form-item :label="$t('business.llmmodels.temperature')" prop="temperature">
          <el-slider v-model="form.temperature" :min="0" :max="2" :step="0.1" show-input />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">{{ $t('commons.button.confirm') }}</el-button>
          <el-button @click="onCancel">{{ $t('commons.button.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { getLLMModel, updateLLMModel } from '@/api/llmmodels';

export default {
  name: 'EditLLMModel',
  data() {
    return {
      loading: false,
      form: {
        name: '',
        baseUri: '',
        modelName: '',
        apiKey: '',
        temperature: 0.7,
        metadata: {
          name: '',
          uuid: '',
          labels: {},
          annotations: {}
        }
      },
      rules: {
        baseUri: [
          { required: true, message: this.$t('commons.validate.input', [this.$t('business.llmmodels.base_uri')]), trigger: 'blur' }
        ],
        modelName: [
          { required: true, message: this.$t('commons.validate.input', [this.$t('business.llmmodels.model_name')]), trigger: 'blur' }
        ]
      }
    };
  },
  created() {
    this.fetchData();
  },
  methods: {
    fetchData() {
      this.loading = true;
      getLLMModel(this.$route.params.name)
        .then(response => {
          this.form = response.data;
          if (!this.form.metadata) {
            this.form.metadata = {
              name: this.form.name,
              uuid: this.form.id || '',
              labels: {},
              annotations: {}
            };
          }
        })
        .catch(error => {
          console.error(error);
          this.$message.error(this.$t('commons.msg.get_failed'));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    onSubmit() {
      this.$refs.form.validate(valid => {
        if (valid) {
          // 确保设置metadata字段
          this.form.metadata.name = this.form.name;
          if (this.form.id && !this.form.metadata.uuid) {
            this.form.metadata.uuid = this.form.id;
          }
          
          updateLLMModel(this.form.name, this.form)
            .then(() => {
              this.$message.success(this.$t('commons.msg.update_success'));
              this.$router.push('/llmmodels');
            })
            .catch(error => {
              console.error(error);
              this.$message.error(error.response?.data?.message || this.$t('commons.msg.update_failed'));
            });
        } else {
          return false;
        }
      });
    },
    onCancel() {
      this.$router.push('/llmmodels');
    }
  }
};
</script>

<style scoped>
.el-form-item-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
  padding-top: 4px;
}
</style> 