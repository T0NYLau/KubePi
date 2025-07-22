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
            <div v-if="hasResponseContent" class="response-message">
              <div class="response-role">{{ getResponseRole() }}:</div>
              <div class="response-text">{{ getResponseContent() }}</div>
            </div>
            <div v-if="response.usage" class="response-usage">
              <p>Tokens: {{ response.usage.prompt_tokens || 0 }} (prompt) + {{ response.usage.completion_tokens || 0 }} (completion) = {{ response.usage.total_tokens || 0 }} (total)</p>
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
  computed: {
    hasResponseContent() {
      // 检查各种可能包含内容的位置
      if (!this.response) return false;
      
      // 检查标准的choices数组
      if (this.response.choices && this.response.choices.length > 0) {
        const choice = this.response.choices[0];
        if (choice.message && choice.message.content) {
          return choice.message.content.trim() !== '';
        }
      }
      
      // 检查其他可能的格式
      if (this.response.assistant && this.response.assistant.trim() !== '') return true;
      if (this.response.text && this.response.text.trim() !== '') return true;
      if (this.response.content && this.response.content.trim() !== '') return true;
      
      return false;
    }
  },
  created() {
    this.fetchData();
  },
  methods: {
    formatContent(content) {
      // 移除可能存在的</think>标记和其前面的内容
      if (!content) return '无法获取AI回答，请检查模型配置和连接';
      
      if (typeof content === 'string') {
        // 处理</think>标记
        const thinkEndIndex = content.indexOf('</think>');
        if (thinkEndIndex !== -1) {
          console.log(`检测到</think>标记，位置: ${thinkEndIndex}`);
          const extractedContent = content.substring(thinkEndIndex + 8).trim();
          console.log(`提取后的内容长度: ${extractedContent.length}`);
          console.log(`提取后的内容前50个字符: ${extractedContent.substring(0, 50)}`);
          content = extractedContent;
        }
        
        // 检查内容是否是JSON字符串
        if (content.trim().startsWith('{') && content.trim().endsWith('}')) {
          try {
            console.log('检测到content可能是JSON字符串，尝试解析');
            const contentObj = JSON.parse(content);
            
            // 检查是否有detail字段，这可能是错误消息
            if (contentObj.detail === 'Not Found') {
              return "API返回错误: 资源未找到。请检查模型配置和连接。BaseURI可能配置错误，确保包含完整路径，例如: http://ip:port/v1/chat/completions";
            } 
            // 检查是否有data字段，这可能是嵌套的响应
            else if (contentObj.data) {
              console.log('从content中提取data字段');
              
              // 如果data包含完整的响应结构
              if (contentObj.data.choices && contentObj.data.choices.length > 0 && 
                  contentObj.data.choices[0].message && 
                  contentObj.data.choices[0].message.content) {
                
                console.log('从content.data.choices[0].message.content中提取内容');
                return this.formatContent(contentObj.data.choices[0].message.content);
              }
              
              // 如果data是字符串，可能是直接的内容
              if (typeof contentObj.data === 'string') {
                console.log('从content.data字符串中提取内容');
                return this.formatContent(contentObj.data);
              }
            }
          } catch (e) {
            console.log('解析content为JSON失败:', e);
          }
        }
      }
      
      return content;
    },
    getResponseRole() {
      // 获取响应角色，默认为assistant
      if (this.response.choices && this.response.choices.length > 0) {
        const choice = this.response.choices[0];
        if (choice.message && choice.message.role) {
          return choice.message.role;
        }
      }
      return 'assistant';
    },
    getResponseContent() {
      // 从各种可能的位置获取响应内容
      if (!this.response) return '';
      
      console.log('处理响应内容:', JSON.stringify(this.response).substring(0, 200) + '...');
      
      // 标准的choices数组
      if (this.response.choices && this.response.choices.length > 0) {
        const choice = this.response.choices[0];
        console.log('处理choice:', JSON.stringify(choice).substring(0, 200) + '...');
        
        if (choice.message) {
          // 首先检查reasoning_content字段
          if (choice.message.reasoning_content) {
            console.log('发现reasoning_content字段');
            return this.formatContent(choice.message.reasoning_content);
          }
          // 然后检查content字段
          if (choice.message.content) {
            console.log('发现content字段');
            return this.formatContent(choice.message.content);
          }
        }
        // 对于非标准格式
        if (choice.text) {
          console.log('发现text字段');
          return this.formatContent(choice.text);
        }
        if (choice.content) {
          console.log('发现content字段(直接在choice中)');
          return this.formatContent(choice.content);
        }
        // 如果choice存在但没有有效内容，尝试将整个choice转为字符串
        if (typeof choice === 'string') {
          console.log('choice是字符串');
          return this.formatContent(choice);
        }
      }
      
      // 其他可能的格式
      if (this.response.assistant) {
        console.log('发现assistant字段');
        return this.formatContent(this.response.assistant);
      }
      if (this.response.text) {
        console.log('发现text字段(直接在response中)');
        return this.formatContent(this.response.text);
      }
      if (this.response.content) {
        console.log('发现content字段(直接在response中)');
        return this.formatContent(this.response.content);
      }
      
      // 如果是字符串响应
      if (typeof this.response === 'string') {
        console.log('响应本身是字符串');
        return this.formatContent(this.response);
      }
      
      // 最后的尝试：将整个响应转为字符串
      try {
        const responseStr = JSON.stringify(this.response);
        if (responseStr && responseStr !== '{}' && responseStr !== '[]') {
          console.log('将整个响应转为字符串');
          return '响应数据结构异常，原始响应: ' + responseStr;
        }
      } catch (e) {
        console.error('无法将响应转为字符串:', e);
      }
      
      console.log('未找到任何内容字段');
      // 如果找不到内容，返回空字符串
      return '无法获取AI回答，请检查模型配置和连接';
    },
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
      
      // 初始化一个临时响应，显示等待消息
      this.response = {
        choices: [{
          message: {
            role: 'assistant',
            content: '正在等待AI响应，大型模型可能需要较长时间...'
          }
        }],
        usage: { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 }
      };
      
      // 记录开始时间
      const startTime = new Date();
      console.log(`开始测试LLM: ${startTime.toISOString()}`);
      
      // 显示加载提示
      const loadingMessage = this.$message({
        message: '正在等待AI响应，这可能需要一段时间...',
        type: 'info',
        duration: 0
      });
      
      testLLMModel(this.modelInfo.name, this.form.content)
        .then(response => {
          // 关闭加载提示
          loadingMessage.close();
          
          console.log('API response:', response);
          console.log('API response data:', response.data);
          const responseTime = new Date() - startTime;
          console.log(`LLM响应时间: ${responseTime}ms`);
          
          // 深度检查响应结构
          this.inspectResponse(response.data);
          
          // 如果response.data存在但没有预期的结构，创建一个兼容的结构
          if (!response.data || (!response.data.choices || response.data.choices.length === 0)) {
            console.log('响应缺少choices数组或为空，创建兼容结构');
            // 处理可能包含答案的其他字段
            let content = '无法提取AI回答';
            
            if (response.data) {
              // 检查各种可能包含内容的字段
              if (response.data.assistantMessage) {
                content = response.data.assistantMessage;
                console.log('从assistantMessage字段提取内容');
              } else if (response.data.response) {
                content = response.data.response;
                console.log('从response字段提取内容');
              } else if (response.data.assistant) {
                content = response.data.assistant;
                console.log('从assistant字段提取内容');
              } else if (response.data.text) {
                content = response.data.text;
                console.log('从text字段提取内容');
              } else if (response.data.content) {
                content = response.data.content;
                console.log('从content字段提取内容');
              } else if (typeof response.data === 'string') {
                // 如果整个响应就是一个字符串
                content = response.data;
                console.log('响应本身是字符串');
              }
            }
            
            // 处理可能的</think>标记
            content = this.formatContent(content);
            
            // 创建兼容的结构
            this.response = {
              choices: [{
                message: {
                  role: 'assistant',
                  content: content
                }
              }],
              usage: (response.data && response.data.usage) || { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 }
            };
          } else {
          this.response = response.data;
            
            // 确保所有content字段都经过formatContent处理
            if (this.response.choices && this.response.choices.length > 0) {
              for (let i = 0; i < this.response.choices.length; i++) {
                if (this.response.choices[i].message && this.response.choices[i].message.content) {
                  this.response.choices[i].message.content = this.formatContent(this.response.choices[i].message.content);
                }
              }
            }
          }
          
          // 打印最终处理后的响应内容
          const finalContent = this.getResponseContent();
          console.log(`最终显示的内容(前100个字符): ${finalContent.substring(0, 100)}`);
          
          this.fetchData();
        })
        .catch(error => {
          // 关闭加载提示
          loadingMessage.close();
          
          console.error('LLM API错误:', error);
          const errorMsg = error.response?.data?.message || error.message || this.$t('commons.msg.operation_failed');
          
          // 如果是超时错误，提供更明确的提示
          const isTimeout = error.code === 'ECONNABORTED' || 
                            errorMsg.includes('timeout') || 
                            errorMsg.includes('超时');
          
          if (isTimeout) {
            this.$message.error('请求超时，请检查LLM服务是否正常运行或增加超时时间');
            this.response = {
              error: {
                message: '请求超时，大型模型可能需要更长的响应时间'
              }
            };
          } else {
            this.$message.error(errorMsg);
          this.response = {
            error: {
                message: errorMsg
            }
          };
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
    
    // 深度检查响应结构
    inspectResponse(data) {
      console.log('检查响应结构:');
      if (!data) {
        console.log('响应数据为空');
        return;
      }
      
      console.log('响应类型:', typeof data);
      if (typeof data === 'object') {
        console.log('顶级字段:', Object.keys(data).join(', '));
        
        if (data.choices && Array.isArray(data.choices)) {
          console.log('choices数组长度:', data.choices.length);
          
          if (data.choices.length > 0) {
            console.log('第一个choice字段:', Object.keys(data.choices[0]).join(', '));
            
            if (data.choices[0].message) {
              console.log('message字段:', Object.keys(data.choices[0].message).join(', '));
              
              if (data.choices[0].message.content) {
                const content = data.choices[0].message.content;
                console.log('content长度:', content.length);
                console.log('content前50个字符:', content.substring(0, 50));
                console.log('content是否包含</think>:', content.includes('</think>'));
                
                if (content.includes('</think>')) {
                  const thinkEndIndex = content.indexOf('</think>');
                  console.log('</think>位置:', thinkEndIndex);
                  console.log('</think>后的内容前50个字符:', content.substring(thinkEndIndex + 8, thinkEndIndex + 58));
                }
              }
              
              if (data.choices[0].message.reasoning_content) {
                console.log('存在reasoning_content字段');
              }
            }
          }
        }
      }
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
  background-color: #1e1e1e;
  color: #ffffff;
  padding: 15px;
  border-radius: 4px;
}
.response-message {
  margin-bottom: 10px;
}
.response-role {
  font-weight: bold;
  margin-bottom: 5px;
  color: #61afef;
}
.response-text {
  white-space: pre-wrap;
  font-family: monospace;
  color: #e6e6e6;
}
.response-usage {
  margin-top: 15px;
  color: #8a8a8a;
  font-size: 12px;
  border-top: 1px dashed #444;
  padding-top: 10px;
}
</style> 