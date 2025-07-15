<template>
  <div class="app-container">
    <el-card class="box-card" shadow="never">
      <div slot="header" class="clearfix">
        <span>{{ $t('business.llmmodels.test_llm_model') }}</span>
      </div>
      <div v-loading="loading">
        <el-form ref="form" :model="form" label-width="120px" size="small">
          <el-form-item :label="$t('business.llmmodels.name')">
            <el-input v-model="modelInfo.name" disabled />
          </el-form-item>
          <el-form-item :label="$t('business.llmmodels.base_uri')">
            <el-input v-model="modelInfo.baseUri" disabled />
          </el-form-item>
          <el-form-item :label="$t('business.llmmodels.model_name')">
            <el-input v-model="modelInfo.modelName" disabled />
          </el-form-item>
          <el-form-item :label="$t('business.llmmodels.status')">
            <el-tag :type="getStatusType(modelInfo.status)">
              {{ $t('business.llmmodels.' + modelInfo.status) }}
            </el-tag>
          </el-form-item>
          <el-form-item :label="$t('business.llmmodels.test_content')" prop="content">
            <el-input 
              type="textarea" 
              v-model="form.content" 
              :rows="4"
              :placeholder="$t('business.llmmodels.test_content')"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="onTest" :disabled="form.content === ''">
              {{ $t('business.llmmodels.test_connection') }}
            </el-button>
            <el-button @click="onBack">{{ $t('commons.button.cancel') }}</el-button>
          </el-form-item>
        </el-form>

        <div v-if="response" class="response-section">
          <h3>{{ $t('business.llmmodels.test_response') }}</h3>
          <el-alert
            v-if="response.error"
            :title="response.error.message"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else class="response-content">
            <div v-for="(choice, index) in response.choices" :key="index" class="response-message">
              <div class="response-role">{{ choice.message.role }}:</div>
              <div class="response-text">{{ choice.message.content }}</div>
            </div>
            <div v-if="response.usage" class="response-usage">
              <p>Tokens: {{ response.usage.prompt_tokens }} (prompt) + {{ response.usage.completion_tokens }} (completion) = {{ response.usage.total_tokens }} (total)</p>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { getLLMModel, testLLMModel } from '@/api/llmmodels';

export default {
  name: 'TestLLMModel',
  data() {
    return {
      loading: false,
      modelInfo: {
        name: '',
        baseUri: '',
        modelName: '',
        status: 'untested'
      },
      form: {
        content: ''
      },
      response: null
    };
  },
  created() {
    this.fetchData();
  },
  methods: {
    getStatusType(status) {
      switch(status) {
        case 'available':
          return 'success';
        case 'unavailable':
          return 'danger';
        case 'untested':
          return 'info';
        default:
          return '';
      }
    },
    fetchData() {
      this.loading = true;
      getLLMModel(this.$route.params.name)
        .then(response => {
          this.modelInfo = response.data;
        })
        .catch(error => {
          console.error(error);
          this.$message.error(this.$t('commons.msg.get_failed'));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    onTest() {
      if (!this.form.content) {
        this.$message.warning(this.$t('commons.validate.input', [this.$t('business.llmmodels.test_content')]));
        return;
      }

      this.loading = true;
      this.response = null;
      
      testLLMModel(this.modelInfo.name, this.form.content)
        .then(response => {
          this.response = response.data;
          this.fetchData(); // Refresh status after test
        })
        .catch(error => {
          console.error(error);
          this.$message.error(error.response?.data?.message || this.$t('commons.msg.operation_failed'));
          this.response = {
            error: {
              message: error.response?.data?.message || this.$t('commons.msg.operation_failed')
            }
          };
        })
        .finally(() => {
          this.loading = false;
        });
    },
    onBack() {
      this.$router.push('/llmmodels');
    }
  }
};
</script>

<style scoped>
.response-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}
.response-content {
  margin-top: 15px;
  background-color: #f8f8f8;
  padding: 15px;
  border-radius: 4px;
}
.response-message {
  margin-bottom: 10px;
}
.response-role {
  font-weight: bold;
  margin-bottom: 5px;
}
.response-text {
  white-space: pre-wrap;
  font-family: monospace;
}
.response-usage {
  margin-top: 15px;
  color: #666;
  font-size: 12px;
  border-top: 1px dashed #ddd;
  padding-top: 10px;
}
</style> 