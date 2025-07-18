<template>
  <div class="app-container">
    <el-card class="box-card" shadow="never">
      <div slot="header" class="clearfix">
        <span>{{ $t('business.llmmodels.create_llm_model') }}</span>
      </div>
      <el-form ref="form" :model="form" :rules="rules" label-width="120px" size="small">
        <el-form-item :label="$t('business.llmmodels.name')" prop="name">
          <el-input v-model="form.name" />
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
          <el-button type="primary" @click="onSubmit">{{ $t('commons.button.create') }}</el-button>
          <el-button @click="onCancel">{{ $t('commons.button.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { createLLMModel } from '@/api/llmmodels';
import { v4 as uuidv4 } from 'uuid';

export default {
  name: 'CreateLLMModel',
  data() {
    return {
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
        name: [
          { required: true, message: this.$t('commons.validate.input', [this.$t('business.llmmodels.name')]), trigger: 'blur' },
          { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: this.$t('commons.validate.name_rules'), trigger: 'blur' }
        ],
        baseUri: [
          { required: true, message: this.$t('commons.validate.input', [this.$t('business.llmmodels.base_uri')]), trigger: 'blur' }
        ],
        modelName: [
          { required: true, message: this.$t('commons.validate.input', [this.$t('business.llmmodels.model_name')]), trigger: 'blur' }
        ]
      }
    };
  },
  methods: {
    onSubmit() {
      this.$refs.form.validate(valid => {
        if (valid) {
          // 确保设置元数据
          this.form.metadata.name = this.form.name;
          
          createLLMModel(this.form)
            .then(() => {
              this.$message.success(this.$t('commons.msg.create_success'));
              this.$router.push('/llmmodels');
            })
            .catch(error => {
              console.error(error);
              this.$message.error(error.response?.data?.message || this.$t('commons.msg.create_failed'));
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