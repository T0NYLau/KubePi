<template>
  <div>
    <div v-loading="loading">
      <!-- 移除表单验证，简化表单结构 -->
      <el-form ref="form" :model="form" label-width="120px" size="small">
        <el-form-item :label="$t('commons.table.name') + ' / ' + $t('business.namespace.namespace')">
          <div>
            <strong>{{ $t('commons.table.name') }}:</strong> {{ podName }}
          </div>
          <div>
            <strong>{{ $t('business.namespace.namespace') }}:</strong> {{ namespace }}
          </div>
          <div>
            <strong>{{ $t('commons.table.status') }}:</strong> {{ podStatus }}
          </div>
        </el-form-item>
        
        <el-form-item :label="$t('business.pod.analysis_scope')">
          <el-checkbox-group v-model="form.analysisTypes">
            <el-checkbox label="status">{{ $t('business.pod.pod_status') }}</el-checkbox>
            <el-checkbox label="logs">{{ $t('business.pod.pod_logs') }}</el-checkbox>
            <el-checkbox label="events">{{ $t('business.pod.pod_events') }}</el-checkbox>
            <el-checkbox label="yaml">yaml</el-checkbox>
            <el-checkbox label="previousLogs">上一次失败的日志</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        
        <el-form-item :label="$t('business.pod.log_lines')" v-if="form.analysisTypes.includes('logs') || form.analysisTypes.includes('previousLogs')">
          <el-select v-model="form.logLines" :placeholder="$t('business.pod.lines')">
            <el-option :label="$t('business.pod.last_20_lines')" value="20"></el-option>
            <el-option :label="$t('business.pod.last_100_lines')" value="100"></el-option>
            <el-option :label="$t('business.pod.last_200_lines')" value="200"></el-option>
            <el-option :label="$t('business.pod.last_500_lines')" value="500"></el-option>
          </el-select>
        </el-form-item>
      </el-form>

      <div v-if="!hasAvailableModels" class="warning-message">
        <el-alert
          :title="$t('business.pod.no_models')"
          type="warning"
          description="请先在系统中配置并测试至少一个可用的LLM模型。"
          show-icon
          :closable="false">
        </el-alert>
      </div>

      <div v-if="models.length > 0" class="model-selection">
        <el-form :model="form" label-width="120px" size="small">
          <el-form-item :label="$t('business.pod.select_model')">
            <el-select v-model="form.selectedModel" :placeholder="$t('business.pod.select_model')">
              <el-option
                v-for="model in availableModels"
                :key="model.id"
                :label="model.name"
                :value="model.name">
              </el-option>
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <div class="button-container">
        <el-button type="primary" @click="onAnalyze" :disabled="analyzing || !hasAvailableModels">
          {{ analyzing ? $t('business.pod.analyzing') : $t('business.pod.analyze') }}
        </el-button>
        <el-button @click="$emit('close')">关闭</el-button>
      </div>

      <div v-if="result || streamingResult" class="analysis-result dark-theme">
        <div class="top-actions">
          <h3>{{ $t('business.pod.analysis_result') }}</h3>
          <el-button type="text" icon="el-icon-top" @click="scrollToTop" class="scroll-top-btn">回到顶部</el-button>
        </div>
        <el-divider></el-divider>
        <div class="result-content">
          <div v-html="formattedResult" v-if="!streamingMode"></div>
          <div v-if="streamingMode" v-html="formattedStreamingResult"></div>
        </div>
      
        <!-- 将聊天界面集成到分析结果区域内 -->
        <div v-if="chatHistory && chatHistory.length > 0" class="chat-history">
          <div v-for="(message, index) in chatHistory" :key="index" :class="['chat-message', message.role]" v-if="index > 1">
            <div class="message-content" v-html="formatChatMessage(message.content)"></div>
          </div>
        </div>
        
        <div class="chat-input">
          <el-input
            v-model="chatMessage"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4 }"
            :placeholder="$t('business.pod.chat_placeholder') || '请输入您的问题...'"
            @keyup.ctrl.enter="sendChatMessage"
          ></el-input>
          <el-button type="primary" @click="sendChatMessage" :disabled="chatSending || !chatMessage.trim()">
            {{ chatSending ? '发送中...' : '发送' }}
          </el-button>
          <el-button type="text" icon="el-icon-top" @click="scrollToDialogTop" class="to-top-btn">TOP</el-button>
        </div>
      </div>
      
      <!-- 删除单独的聊天界面区域 -->
    </div>
  </div>
</template>

<script>
import { getWorkLoadByName } from "@/api/workloads"
import { listEventsWithPodSelector } from "@/api/events"
import { getPodLogsByName } from "@/api/pods"
import { getLLMModels, testLLMModel } from "@/api/llmmodels"
import axios from "axios"

// 不使用import，改用全局变量方式
// Vue 2中可能已经全局注册了marked
var marked = window.marked || function(text) { return text; };

export default {
  name: "PodAIAnalysis",
  props: {
    name: {
      type: String,
      required: true
    },
    namespace: {
      type: String,
      required: true
    },
    cluster: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      loading: false,
      analyzing: false,
      podName: this.name,
      namespace: this.namespace,
      clusterName: this.cluster,
      podStatus: "",
      podDetails: null,
      events: [],
      logs: "",
      models: [],
      result: null,
      streamingResult: "",
      streamingMode: true,
      abortController: null,
      // 聊天相关数据
      chatHistory: [],
      chatMessage: "",
      chatSending: false,
      form: {
        analysisTypes: ["status", "events", "yaml", "previousLogs"],
        logLines: "100",
        selectedModel: ""
      }
    }
  },
  computed: {
    availableModels() {
      return this.models.filter(model => model.status === "available");
    },
    hasAvailableModels() {
      return this.availableModels.length > 0;
    },
    formattedResult() {
      if (!this.result) return '';
      try {
        console.log("开始格式化结果...");
        
        // 尝试使用marked处理markdown
        let markedContent = '';
        let markedSuccess = false;
        
        if (typeof marked === 'function') {
          try {
            markedContent = marked(this.result);
            markedSuccess = true;
            console.log("使用marked函数处理成功");
          } catch (e) {
            console.error("使用marked函数处理失败:", e);
          }
        } else if (marked && typeof marked.parse === 'function') {
          try {
            markedContent = marked.parse(this.result);
            markedSuccess = true;
            console.log("使用marked.parse处理成功");
          } catch (e) {
            console.error("使用marked.parse处理失败:", e);
          }
        }
        
        if (!markedSuccess) {
          console.log("使用简单Markdown处理");
          markedContent = this.simpleMarkdownToHtml(this.result);
        }
        
        // 如果有DOMPurify可用，使用它清理HTML
        let finalContent = markedContent;
        if (window.DOMPurify) {
          try {
            finalContent = window.DOMPurify.sanitize(markedContent);
            console.log("使用DOMPurify清理成功");
          } catch (e) {
            console.error("使用DOMPurify清理失败:", e);
          }
        }
        
        // 确保内容不为空
        if (!finalContent || finalContent.trim() === '') {
          console.warn("格式化后的内容为空，使用原始内容");
          return this.result;
        }
        
        console.log("格式化完成");
        return finalContent;
      } catch (e) {
        console.error('处理Markdown出错:', e);
        // 如果处理失败，至少做一些基本的格式化
        try {
          return this.simpleMarkdownToHtml(this.result);
        } catch (e2) {
          console.error('简单Markdown处理也失败:', e2);
          return this.result; // 最后的降级处理
        }
      }
    },
    formattedStreamingResult() {
      if (!this.streamingResult) return '';
      
      try {
        // 简单清理文本，移除可能的元数据
        const cleanText = this.streamingResult.replace(/data:\s*\{.*?\}/g, '');
        
        // 优先使用marked.js进行Markdown渲染
        if (typeof marked === 'function') {
          try {
            return marked(cleanText);
          } catch (e) {
            console.error('marked渲染失败', e);
          }
        } else if (marked && typeof marked.parse === 'function') {
          try {
            return marked.parse(cleanText);
          } catch (e) {
            console.error('marked.parse渲染失败', e);
          }
        }
        
        // 如果marked不可用或失败，使用简单替换
        return cleanText
          .replace(/\n/g, '<br>')
          .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
          .replace(/\*([^*]+)\*/g, '<em>$1</em>')
          .replace(/`([^`]+)`/g, '<code>$1</code>');
          } catch (e) {
        console.error('格式化流式内容失败:', e);
        return this.streamingResult;
      }
    }
  },
  // 使用mounted而不是created，确保DOM已经准备好
  mounted() {
    this.fetchPodDetails();
    this.fetchModels();
  },
  methods: {
    async fetchPodDetails() {
      this.loading = true;
      try {
        // 获取Pod详情
        const podDetails = await getWorkLoadByName(this.clusterName, "pods", this.namespace, this.podName);
        this.podDetails = podDetails;
        this.podStatus = podDetails.status.phase || "Unknown";
      } catch (error) {
        console.error("获取Pod详情失败:", error);
        this.$message.error("获取Pod详情失败");
      } finally {
        this.loading = false;
      }
    },
    
    // 简单的Markdown转HTML处理
    simpleMarkdownToHtml(markdownText) {
      if (!markdownText) return '';
      
      // 清理并统一换行符
      let processedMarkdown = markdownText
        .replace(/\r\n/g, '\n')  // 统一换行符
        .replace(/\u00A0/g, ' ') // 替换不间断空格
        .replace(/chatcmpl-[a-zA-Z0-9]+/g, '') // 移除chatcmpl-ID
        .replace(/\b[a-f0-9]{32}\b/g, '') // 移除32位十六进制ID
        .replace(/\b[a-f0-9]{24}\b/g, '') // 移除24位十六进制ID
        .replace(/\b[a-f0-9]{8}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{12}\b/g, '') // 移除UUID格式
        .replace(/\b[0-9a-f]{8,40}\b/g, ''); // 移除其他可能的哈希/ID格式
      
      // 处理代码块 (必须先处理，避免内部内容被其他规则匹配)
      processedMarkdown = processedMarkdown.replace(/```([\s\S]*?)```/g, function(match, code) {
        return '<pre><code>' + code.replace(/</g, '&lt;').replace(/>/g, '&gt;') + '</code></pre>';
      });
      
      // 处理未闭合的代码块 (流式输出可能导致代码块未闭合)
      const unclosedCodeBlockMatch = processedMarkdown.match(/```([\s\S]*)$/);
      if (unclosedCodeBlockMatch) {
        const code = unclosedCodeBlockMatch[1];
        processedMarkdown = processedMarkdown.replace(/```([\s\S]*)$/, '<pre><code>' + code.replace(/</g, '&lt;').replace(/>/g, '&gt;') + '</code></pre>');
      }
      
      // 处理内联代码 (必须在代码块之后处理)
      processedMarkdown = processedMarkdown.replace(/`([^`]+)`/g, function(match, code) {
        return '<code>' + code.replace(/</g, '&lt;').replace(/>/g, '&gt;') + '</code>';
      });
      
      // 处理标题
      processedMarkdown = processedMarkdown
        .replace(/^### (.*$)/gim, '<h3>$1</h3>')
        .replace(/^## (.*$)/gim, '<h2>$1</h2>')
        .replace(/^# (.*$)/gim, '<h1>$1</h1>');
      
      // 改进列表处理
      // 处理有序列表
      processedMarkdown = processedMarkdown.replace(/^\d+\.\s+(.*)$/gm, '<li>$1</li>');
      
      // 处理无序列表 (改进列表处理，避免嵌套问题)
      const listItemRegex = /^[*\-+] (.*)$/gm;
      let listMatches;
      let lastIndex = 0;
      let result = '';
      
      while ((listMatches = listItemRegex.exec(processedMarkdown)) !== null) {
        // 添加匹配前的内容
        result += processedMarkdown.substring(lastIndex, listMatches.index);
        
        // 添加列表项
        result += '<li>' + listMatches[1] + '</li>';
        
        // 更新最后处理位置
        lastIndex = listMatches.index + listMatches[0].length;
      }
      
      // 添加剩余内容
      result += processedMarkdown.substring(lastIndex);
      processedMarkdown = result;
      
      // 检查是否有列表项，如果有则添加<ul>标签
      if (processedMarkdown.includes('<li>')) {
        // 改进列表处理逻辑，处理多个列表的情况
        const segments = processedMarkdown.split(/<\/li>\s*(?![^\n]*<li>)/g);
        processedMarkdown = '';
        
        for (let i = 0; i < segments.length; i++) {
          const segment = segments[i];
          if (segment.includes('<li>')) {
            // 找到列表开始和结束位置
            const startIdx = segment.indexOf('<li>');
            const endIdx = segment.lastIndexOf('</li>') + 5;
            
            if (startIdx >= 0 && endIdx >= 5) {
              // 将列表包装在<ul>标签中
              processedMarkdown += 
                segment.substring(0, startIdx) + 
                '<ul>' + 
                segment.substring(startIdx, endIdx) + 
                '</ul>' + 
                segment.substring(endIdx);
            } else {
              processedMarkdown += segment;
            }
          } else {
            processedMarkdown += segment;
          }
        }
      }
      
      // 处理表格
      const tableRegex = /^\|(.+)\|$/gm;
      if (tableRegex.test(processedMarkdown)) {
        // 找到所有表格行
        const tableRows = processedMarkdown.match(/^\|(.+)\|$/gm);
        if (tableRows && tableRows.length > 1) {
          let tableHtml = '<table class="markdown-table">';
          
          // 处理表头
          const headerRow = tableRows[0];
          const headerCells = headerRow.split('|').filter(cell => cell.trim() !== '');
          tableHtml += '<thead><tr>';
          for (const cell of headerCells) {
            tableHtml += `<th>${cell.trim()}</th>`;
          }
          tableHtml += '</tr></thead>';
          
          // 跳过分隔行
          const dataRows = tableRows.slice(2);
          if (dataRows.length > 0) {
            tableHtml += '<tbody>';
            for (const row of dataRows) {
              const cells = row.split('|').filter(cell => cell.trim() !== '');
              tableHtml += '<tr>';
              for (const cell of cells) {
                tableHtml += `<td>${cell.trim()}</td>`;
              }
              tableHtml += '</tr>';
            }
            tableHtml += '</tbody>';
          }
          
          tableHtml += '</table>';
          
          // 替换原始表格文本
          for (const row of tableRows) {
            processedMarkdown = processedMarkdown.replace(row, '');
          }
          
          // 插入HTML表格
          processedMarkdown = tableHtml + processedMarkdown;
        }
      }
      
      // 处理加粗
      processedMarkdown = processedMarkdown.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
      
      // 处理斜体
      processedMarkdown = processedMarkdown.replace(/\*([^*]+)\*/g, '<em>$1</em>');
      
      // 处理链接
      processedMarkdown = processedMarkdown.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>');
      
      // 处理段落和换行
      processedMarkdown = processedMarkdown
        .replace(/\n\n/g, '</p><p>') // 空行分隔段落
        .replace(/\n/g, '<br />');   // 单个换行
      
      // 确保内容被包裹在段落标签中
      if (!processedMarkdown.startsWith('<h1>') && 
          !processedMarkdown.startsWith('<h2>') && 
          !processedMarkdown.startsWith('<h3>') && 
          !processedMarkdown.startsWith('<p>') &&
          !processedMarkdown.startsWith('<ul>') &&
          !processedMarkdown.startsWith('<table>')) {
        processedMarkdown = '<p>' + processedMarkdown;
      }
      
      if (!processedMarkdown.endsWith('</p>') && 
          !processedMarkdown.endsWith('</h1>') && 
          !processedMarkdown.endsWith('</h2>') && 
          !processedMarkdown.endsWith('</h3>') && 
          !processedMarkdown.endsWith('</ul>') &&
          !processedMarkdown.endsWith('</table>')) {
        processedMarkdown += '</p>';
      }
      
      return processedMarkdown;
    },
    
    // 查找content字段并增强调试
    findContentInResponse(obj, depth = 0, maxDepth = 5) {
      // 防止无限递归
      if (depth > maxDepth) return null;
      
      // 如果是null或undefined，直接返回null
      if (obj === null || obj === undefined) return null;
      
      // 打印调试信息
      if (depth === 0) {
        console.log("开始解析响应内容:", typeof obj, obj ? JSON.stringify(obj).substring(0, 100) + "..." : "null");
      }
      
      // 如果是字符串，尝试解析JSON
      if (typeof obj === 'string') {
        try {
          return this.findContentInResponse(JSON.parse(obj), depth + 1, maxDepth);
        } catch (e) {
          // 如果不是有效的JSON，直接返回字符串
          return obj;
        }
      }
      
      // 如果找到message.content字段
      if (obj.message && obj.message.content) {
        console.log("找到message.content字段:", obj.message.content.substring(0, 50) + "...");
        return obj.message.content;
      }
      
      // 如果是数组，遍历数组
      if (Array.isArray(obj)) {
        for (const item of obj) {
          const content = this.findContentInResponse(item, depth + 1, maxDepth);
          if (content) return content;
        }
        return null;
      }
      
      // 如果是对象，遍历对象的属性
      if (typeof obj === 'object') {
        // 优先检查常见的路径
        if (obj.choices && Array.isArray(obj.choices) && obj.choices.length > 0) {
          const choice = obj.choices[0];
          if (choice.delta && choice.delta.content) {
            console.log("找到choice.delta.content字段:", choice.delta.content);
            return choice.delta.content;
          }
          if (choice.message && choice.message.content) {
            console.log("找到choice.message.content字段:", choice.message.content.substring(0, 50) + "...");
            return choice.message.content;
          }
          if (choice.text) {
            console.log("找到choice.text字段:", choice.text.substring(0, 50) + "...");
            return choice.text;
          }
          
          const content = this.findContentInResponse(choice, depth + 1, maxDepth);
          if (content) return content;
        }
        
        if (obj.data) {
          const content = this.findContentInResponse(obj.data, depth + 1, maxDepth);
          if (content) return content;
        }
        
        // 遍历所有属性
        for (const key in obj) {
          if (key === 'content' && typeof obj[key] === 'string') {
            console.log("找到content字段:", obj[key].substring(0, 50) + "...");
            return obj[key];
          }
          
          if (key === 'text' && typeof obj[key] === 'string') {
            console.log("找到text字段:", obj[key].substring(0, 50) + "...");
            return obj[key];
          }
          
          const content = this.findContentInResponse(obj[key], depth + 1, maxDepth);
          if (content) return content;
        }
      }
      
      return null;
    },
    
    async fetchModels() {
      try {
        const response = await getLLMModels();
        if (response && response.data) {
          this.models = response.data;
          if (this.availableModels.length > 0) {
            this.form.selectedModel = this.availableModels[0].name;
          }
        }
      } catch (error) {
        console.error("获取LLM模型列表失败:", error);
        this.$message.error("获取LLM模型列表失败");
      }
    },
    async fetchPodEvents() {
      try {
        if (!this.podDetails || !this.podDetails.metadata || !this.podDetails.metadata.uid) {
          console.warn("Pod详情不完整，无法获取事件");
          return [];
        }
        
        let selects = "involvedObject.name={PodName}&involvedObject.namespace={Namespace}&involvedObject.uid={Uid}&limit=50";
        selects = selects
          .replace("{PodName}", this.podName)
          .replace("{Namespace}", this.namespace)
          .replace("{Uid}", this.podDetails.metadata.uid);
        
        const response = await listEventsWithPodSelector(this.clusterName, this.namespace, selects);
        return response && response.items ? response.items : [];
      } catch (error) {
        console.error("获取Pod事件失败:", error);
        this.$message.error("获取Pod事件失败");
        return [];
      }
    },
    async fetchPodLogs() {
      try {
        if (!this.podDetails || !this.podDetails.spec || !this.podDetails.spec.containers) {
          console.warn("Pod详情不完整，无法获取日志");
          return "";
        }
        
        const params = {
          tailLines: parseInt(this.form.logLines) || 100,
          timestamps: true
        };
        
        // 如果Pod有多个容器，获取第一个容器的日志
        if (this.podDetails.spec.containers.length > 0) {
          params.container = this.podDetails.spec.containers[0].name;
        }
        
        console.log("获取Pod日志，参数:", params);
        const response = await getPodLogsByName(this.clusterName, this.namespace, this.podName, params);
        console.log("Pod日志响应:", response);
        if (response && typeof response === 'object' && response.data) {
          return response.data;
        } else if (typeof response === 'string') {
          return response;
        } else {
          console.warn("获取到的日志格式不正确:", response);
          return "";
        }
      } catch (error) {
        console.error("获取Pod日志失败:", error);
        this.$message.error("获取Pod日志失败: " + (error.message || "未知错误"));
        return "";
      }
    },
    // 处理转义字符和特殊字符
    processEscapeCharacters(text) {
      if (!text) return '';
      
      // 处理常见的转义字符
      return text
        .replace(/\\n/g, '\n')    // 换行符
        .replace(/\\r/g, '\r')    // 回车符
        .replace(/\\t/g, '\t')    // 制表符
        .replace(/\\"/g, '"')     // 双引号
        .replace(/\\'/g, "'")     // 单引号
        .replace(/\\\\/g, '\\')   // 反斜杠
        .replace(/\\b/g, '\b')    // 退格符
        .replace(/\\f/g, '\f');   // 换页符
    },
    
    // 处理流式显示
    async streamData(prompt) {
      try {
        this.streamingResult = '';
        this.abortController = new AbortController();
        const signal = this.abortController.signal;
        this.$nextTick(() => {
          const resultContent = document.querySelector('.result-content');
          if (resultContent) {
            resultContent.scrollTop = resultContent.scrollHeight;
          }
        });
        // 只拼接content内容，过滤无用信息
        const handleStreamContent = (text) => {
          try {
            if (!text || !text.trim()) return;
            
            // 跳过纯ID（如 -chatcmpl-xxxx）
            if (/^-?chatcmpl-[a-zA-Z0-9]+$/.test(text.trim())) return;
            
            // 跳过其他形式的ID/哈希值
            if (/^\s*\b[a-f0-9]{24,40}\b\s*$/.test(text.trim())) return;
            if (/^\s*\b[a-f0-9]{8}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{12}\b\s*$/.test(text.trim())) return;
            
            // 跳过完整JSON但没有content字段
            if (text.startsWith('{') && text.endsWith('}')) {
              try {
                const json = JSON.parse(text);
                if (json.choices && json.choices[0] && json.choices[0].delta && typeof json.choices[0].delta.content === 'string') {
                  text = json.choices[0].delta.content;
                } else if (json.choices && json.choices[0] && json.choices[0].message && typeof json.choices[0].message.content === 'string') {
                  text = json.choices[0].message.content;
                } else if (typeof json.content === 'string') {
                  text = json.content;
                } else {
                  return;
                }
              } catch (e) {
                return;
              }
            }
            if (text && text.trim()) {
              this.$set(this, 'streamingResult', this.streamingResult + text);
              this.analyzing = false;
            this.$nextTick(() => {
                const resultContent = document.querySelector('.result-content');
                if (resultContent) {
                  resultContent.scrollTop = resultContent.scrollHeight;
                }
            });
            }
          } catch (e) {
            console.error('处理流数据出错:', e);
          }
        };
        console.log("开始流式请求...");
        const response = await testLLMModel(this.form.selectedModel, prompt, {
          streaming: true,
          onProgress: handleStreamContent,
          signal: signal
        });
        console.log("流式请求完成");
        if (this.streamingResult) {
            this.result = this.streamingResult;
            this.chatHistory = [
              { role: 'user', content: prompt },
            { role: 'assistant', content: this.streamingResult }
            ];
          }
        return this.streamingResult;
      } catch (error) {
        if (axios.isCancel(error)) {
          console.log('请求被取消');
        } else {
          console.error('流式请求失败:', error);
          this.$message.error("分析失败: " + (error.message || "未知错误"));
        }
      } finally {
        this.analyzing = false;
        this.abortController = null;
      }
    },
    
    // 取消正在进行的流式请求
    cancelStream() {
      if (this.abortController) {
        this.abortController.abort();
        this.abortController = null;
      }
    },
    
    // 格式化聊天消息
    formatChatMessage(message) {
      try {
        if (!message) return '';
        
        // 首先过滤掉各种ID
        message = message.replace(/chatcmpl-[a-zA-Z0-9]+/g, '')
                         .replace(/\b[a-f0-9]{32}\b/g, '')  // 移除32位十六进制ID
                         .replace(/\b[a-f0-9]{24}\b/g, '')  // 移除24位十六进制ID
                         .replace(/\b[a-f0-9]{8}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{4}[a-f0-9]{12}\b/g, '') // 移除UUID格式
                         .replace(/\b[0-9a-f]{8,40}\b/g, '') // 移除其他可能的哈希/ID格式
                         .replace(/^-+$/gm, ''); // 移除纯短横线行
        
        // 尝试使用marked处理markdown
        let markedContent = '';
        let markedSuccess = false;
        
        // 检查是否有marked可用
        if (window.marked) {
          try {
            if (typeof window.marked === 'function') {
              markedContent = window.marked(message);
              markedSuccess = true;
            } else if (window.marked.parse) {
              markedContent = window.marked.parse(message);
              markedSuccess = true;
            }
          } catch (e) {
            console.error("使用marked处理markdown失败:", e);
          }
        }
        
        // 如果marked不可用或失败，使用简单Markdown处理
        if (!markedSuccess) {
          markedContent = this.simpleMarkdownToHtml(message);
        }
        
        // 如果有DOMPurify可用，使用它清理HTML
        let finalContent = markedContent;
        if (window.DOMPurify) {
          try {
            finalContent = window.DOMPurify.sanitize(markedContent);
          } catch (e) {
            console.error("使用DOMPurify清理失败:", e);
          }
        }
        
        return finalContent || message;
      } catch (e) {
        console.error('处理聊天消息出错:', e);
        return message;
      }
    },
    
    // 发送聊天消息
    async sendChatMessage() {
      if (!this.chatMessage.trim() || this.chatSending) return;
      
      const userMessage = this.chatMessage.trim();
      this.chatMessage = '';
      this.chatSending = true;
      
      // 添加用户消息到聊天历史
      this.chatHistory.push({ role: 'user', content: userMessage });
      
      try {
        // 构建聊天上下文
        const chatContext = this.buildChatContext(userMessage);
        
        // 滚动到底部
        this.$nextTick(() => {
          if (this.$refs.chatMessages) {
            this.$refs.chatMessages.scrollTop = this.$refs.chatMessages.scrollHeight;
          }
        });
        
        // 添加临时的助手消息（显示加载中）
        const assistantIndex = this.chatHistory.length;
        this.chatHistory.push({ role: 'assistant', content: '思考中...' });
        
        // 流式请求
        let assistantResponse = '';
        
        // 创建一个新的AbortController
        this.abortController = new AbortController();
        const signal = this.abortController.signal;
        
        // 处理流式响应
        const handleChatProgress = (chunk) => {
          if (!chunk) return;
          
          let content = '';
          
          // 处理不同格式的响应块
          if (typeof chunk === 'string') {
            // 跳过纯ID（如 chatcmpl-xxxx）
            if (/^-?chatcmpl-[a-zA-Z0-9]+$/.test(chunk.trim())) {
              return;
            }
            
            // 尝试处理可能的SSE格式 (data: {...})
            const lines = chunk.split('\n').filter(line => line.trim() !== '');
            
            for (const line of lines) {
              try {
                // 跳过纯ID行
                if (/^-?chatcmpl-[a-zA-Z0-9]+$/.test(line.trim())) {
                  continue;
                }
                
                if (line.startsWith('data: ')) {
                  const jsonStr = line.substring(6).trim();
                  if (jsonStr === '[DONE]') continue;
                  
                  const json = JSON.parse(jsonStr);
                  const chunkContent = this.findContentInResponse(json);
                  if (chunkContent) {
                    content += chunkContent;
                  }
                } else {
                  // 尝试解析为JSON
                  const json = JSON.parse(line);
                  const chunkContent = this.findContentInResponse(json);
                  if (chunkContent) {
                    content += chunkContent;
                  }
                }
              } catch (e) {
                // 如果不是JSON，视为纯文本增量
                // 过滤掉chatcmpl-ID
                if (!/^-?chatcmpl-[a-zA-Z0-9]+$/.test(line.trim())) {
                  content += line;
                }
              }
            }
          } else if (typeof chunk === 'object') {
            const chunkContent = this.findContentInResponse(chunk);
            if (chunkContent) {
              content = chunkContent;
            }
          }
          
          // 如果解析出内容，更新assistantResponse并更新聊天历史
          if (content) {
            content = this.processEscapeCharacters(content);
            assistantResponse += content;
            
            // 更新当前助手消息，使用Vue响应式API
            if (this.chatHistory[assistantIndex]) {
              this.$set(this.chatHistory[assistantIndex], 'content', assistantResponse);
              
              // 强制重新渲染视图
              this.$nextTick(() => {
                // 滚动到底部
                if (this.$refs.chatMessages) {
                  this.$refs.chatMessages.scrollTop = this.$refs.chatMessages.scrollHeight;
                }
              });
            }
          }
        };
        
        // 发送请求
        const response = await testLLMModel(this.form.selectedModel, chatContext, {
          streaming: true,
          onProgress: handleChatProgress,
          signal: signal
        });
        
        // 请求完成后，获取最终结果
        if (response && response.data) {
          const content = this.findContentInResponse(response.data);
          if (content) {
            const processedContent = this.processEscapeCharacters(content);
            if (!assistantResponse || assistantResponse.trim() === '') {
              assistantResponse = processedContent;
            }
            
            // 更新最终的助手回复
            if (this.chatHistory[assistantIndex]) {
              this.$set(this.chatHistory[assistantIndex], 'content', assistantResponse);
            }
          }
        }
      } catch (error) {
        if (axios.isCancel(error)) {
          console.log('聊天请求被取消');
        } else {
          console.error('聊天请求失败:', error);
          this.$message.error("聊天请求失败: " + (error.message || "未知错误"));
          
          // 将错误消息添加到聊天历史
          const lastIndex = this.chatHistory.length - 1;
          if (this.chatHistory[lastIndex] && this.chatHistory[lastIndex].role === 'assistant') {
            this.$set(this.chatHistory[lastIndex], 'content', "抱歉，处理您的请求时出错了。");
          }
        }
      } finally {
        this.chatSending = false;
        this.abortController = null;
        
        // 滚动到底部
        this.$nextTick(() => {
          if (this.$refs.chatMessages) {
            this.$refs.chatMessages.scrollTop = this.$refs.chatMessages.scrollHeight;
          }
        });
      }
    },
    
    // 构建聊天上下文
    buildChatContext(userMessage) {
      // 获取之前的聊天历史
      const previousMessages = this.chatHistory.map(msg => `${msg.role === 'user' ? '用户' : '助手'}: ${msg.content}`).join('\n\n');
      
      // 构建上下文
      let context = `你是Kubernetes专家，请根据以下Pod的信息和聊天历史，回答用户的问题。\n\n`;
      context += `Pod名称: ${this.podName}\n`;
      context += `命名空间: ${this.namespace}\n`;
      context += `当前状态: ${this.podStatus}\n\n`;
      
      // 添加聊天历史
      if (previousMessages) {
        context += `## 聊天历史\n${previousMessages}\n\n`;
      }
      
      // 添加用户最新的问题
      context += `## 用户最新问题\n${userMessage}\n\n`;
      
      return context;
    },
    
    async onAnalyze() {
      if (!this.form.selectedModel) {
        this.$message.warning("请选择一个可用的LLM模型");
        return;
      }

      if (!this.form.analysisTypes || this.form.analysisTypes.length === 0) {
        this.$message.warning("请至少选择一项分析范围");
        return;
      }

      // 重置状态
      this.analyzing = true;
      this.result = null;
      this.streamingResult = '';
      this.chatHistory = []; // 清空聊天历史
      
      try {
        let analysisData = {
          podDetails: this.podDetails,
          events: [],
          logs: "",
          previousLogs: ""
        };
        
        console.log("分析范围:", this.form.analysisTypes);
        
        // 根据选择的分析类型获取相应数据
        const promises = [];
        
        if (this.form.analysisTypes.includes("events")) {
          promises.push(this.fetchPodEvents().then(events => {
            analysisData.events = events;
            console.log("已获取Pod事件数:", events.length);
          }));
        }
        
        if (this.form.analysisTypes.includes("logs")) {
          promises.push(this.fetchPodLogs().then(logs => {
            analysisData.logs = logs;
            console.log("已获取Pod日志长度:", logs ? logs.length : 0);
          }));
        }

        if (this.form.analysisTypes.includes("previousLogs")) {
          promises.push(this.fetchPreviousLogs().then(logs => {
            analysisData.previousLogs = logs;
            console.log("已获取Pod上次失败日志长度:", logs ? logs.length : 0);
          }));
        }
        
        // 等待所有数据获取完成
        await Promise.all(promises);
        
        // 构建发送给AI的提示
        const prompt = this.buildPrompt(analysisData);
        
        console.log("发送AI分析请求，模型:", this.form.selectedModel);
        
        // 立即显示结果区域，准备接收流式内容
        if (this.streamingMode) {
          // 使用流式显示
          await this.streamData(prompt);
        } else {
          // 调用AI进行分析 (非流式方式)
          const response = await testLLMModel(this.form.selectedModel, prompt);
          
          console.log("AI分析响应:", response);
          
          // 检查响应格式
          if (!response) {
            throw new Error("未收到响应");
          }
          
          // 使用原有函数提取内容
          let content = this.findContentInResponse(response.data);
          
          if (!content) {
            // 记录完整的响应结构，帮助调试
            console.error("无法从响应中提取内容，完整响应:", JSON.stringify(response));
            throw new Error("无法从响应中提取内容");
          }
          
          // 处理转义字符
          content = this.processEscapeCharacters(content);
          
          // 检查处理后的内容是否为空
          if (!content || content.trim() === '') {
            throw new Error("处理后的内容为空");
          }
          
          console.log("处理后的AI响应内容:", content.substring(0, 100) + "...");
          this.result = content;
          
          // 分析完成，更新状态
          this.analyzing = false;
        }
      } catch (error) {
        console.error("AI分析失败:", error);
        this.$message.error("AI分析失败: " + (error.message || "未知错误"));
        
        // 确保状态被更新
        this.analyzing = false;
      }
    },
    
    // 获取Pod上一次失败的日志
    async fetchPreviousLogs() {
      try {
        if (!this.podDetails || !this.podDetails.spec || !this.podDetails.spec.containers) {
          console.warn("Pod详情不完整，无法获取上一次失败日志");
          return "";
        }
        
        const params = {
          tailLines: parseInt(this.form.logLines) || 100,
          timestamps: true,
          previous: true // 获取上一次容器的日志
        };
        
        // 如果Pod有多个容器，获取第一个容器的日志
        if (this.podDetails.spec.containers.length > 0) {
          params.container = this.podDetails.spec.containers[0].name;
        }
        
        console.log("获取Pod上一次失败日志，参数:", params);
        const response = await getPodLogsByName(this.clusterName, this.namespace, this.podName, params);
        console.log("Pod上一次失败日志响应:", response);
        if (response && typeof response === 'object' && response.data) {
          return response.data;
        } else if (typeof response === 'string') {
          return response;
        } else {
          console.warn("获取到的上一次失败日志格式不正确:", response);
          return "";
        }
      } catch (error) {
        console.error("获取Pod上一次失败日志失败:", error);
        return "获取上一次失败日志时出错: " + (error.message || "未知错误");
      }
    },
    
    buildPrompt(data) {
      let prompt = `作为Kubernetes专家，请分析以下Pod的信息，找出可能的问题并提供解决方案。

Pod名称: ${this.podName}
命名空间: ${this.namespace}
当前状态: ${this.podStatus}

`;

      // 添加Pod详情
      if (data.podDetails && this.form.analysisTypes.includes("status")) {
        prompt += `\n## Pod详情\n\`\`\`yaml\n`;
        // 添加关键的Pod信息，避免提示过长
        if (data.podDetails.status) {
          prompt += `状态阶段: ${data.podDetails.status.phase || "Unknown"}\n`;
          
          if (data.podDetails.status.conditions) {
            prompt += `状态条件:\n`;
            data.podDetails.status.conditions.forEach(condition => {
              prompt += `  - 类型: ${condition.type}, 状态: ${condition.status}, 原因: ${condition.reason || "无"}\n`;
            });
          }
          
          if (data.podDetails.status.containerStatuses) {
            prompt += `容器状态:\n`;
            data.podDetails.status.containerStatuses.forEach(status => {
              prompt += `  - 容器: ${status.name}, 就绪: ${status.ready}, 重启次数: ${status.restartCount}\n`;
              
              if (status.state) {
                if (status.state.waiting) {
                  prompt += `    等待原因: ${status.state.waiting.reason || "无"}, 信息: ${status.state.waiting.message || "无"}\n`;
                }
                if (status.state.terminated) {
                  prompt += `    终止原因: ${status.state.terminated.reason || "无"}, 退出码: ${status.state.terminated.exitCode}\n`;
                }
              }
            });
          }
        }
        prompt += `\`\`\`\n`;
      }

      // 添加事件信息
      if (data.events && data.events.length > 0 && this.form.analysisTypes.includes("events")) {
        prompt += `\n## Pod事件 (最近${data.events.length}条)\n\`\`\`\n`;
        data.events.forEach(event => {
          const time = event.firstTimestamp || event.eventTime || event.lastTimestamp;
          prompt += `[${time}] 类型: ${event.type}, 原因: ${event.reason}, 信息: ${event.message}\n`;
        });
        prompt += `\`\`\`\n`;
      }

      // 添加日志信息
      if (data.logs && this.form.analysisTypes.includes("logs")) {
        prompt += `\n## Pod日志 (最近${this.form.logLines}行)\n\`\`\`\n${data.logs}\n\`\`\`\n`;
      }

      // 添加上一次失败的日志信息
      if (data.previousLogs && this.form.analysisTypes.includes("previousLogs")) {
        prompt += `\n## Pod上一次失败的日志 (最近${this.form.logLines}行)\n\`\`\`\n${data.previousLogs}\n\`\`\`\n`;
      }

      // 添加YAML信息
      if (data.podDetails && this.form.analysisTypes.includes("yaml")) {
        prompt += `\n## Pod YAML\n\`\`\`yaml\n`;
        try {
          // 安全地序列化YAML
          const yamlData = JSON.stringify(data.podDetails, null, 2);
          prompt += yamlData;
        } catch (e) {
          console.error("YAML序列化失败:", e);
          prompt += "YAML序列化失败";
        }
        prompt += `\n\`\`\`\n`;
      }

      prompt += `\n请提供以下内容：
1. 问题诊断：根据提供的信息，分析Pod可能存在的问题
2. 解决方案：提供具体的解决步骤
3. 最佳实践建议：如何避免类似问题再次发生

请以Markdown格式输出，使用标题、列表和代码块使结果更易读。`;

      return prompt;
    },
    // 滚动到弹窗顶部
    scrollToTop() {
      // 滚动到分析弹窗最顶部
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
      
      // 滚动结果区域到顶部
      const resultContent = document.querySelector('.result-content');
      if (resultContent) {
        resultContent.scrollTop = 0;
      }
    },
    // 滚动到聊天输入框顶部
    scrollToDialogTop() {
      // 尝试多种可能的父元素选择器
      const possibleSelectors = [
        '.el-dialog__body',
        '.el-dialog__wrapper',
        '.el-dialog',
        '.app-container',
        '.ai-analysis-dialog'
      ];
      
      // 尝试找到可滚动的父元素
      let scrollableParent = null;
      
      // 首先尝试获取当前组件的父级元素
      let currentElement = this.$el;
      while (currentElement && !scrollableParent) {
        // 检查当前元素是否可滚动
        if (currentElement.scrollHeight > currentElement.clientHeight) {
          scrollableParent = currentElement;
          console.log('找到可滚动的父元素:', currentElement);
          break;
        }
        
        // 向上查找父元素
        currentElement = currentElement.parentElement;
        
        // 避免无限循环
        if (currentElement === document.body) break;
      }
      
      // 如果通过DOM遍历没找到，尝试通过选择器查找
      if (!scrollableParent) {
        for (const selector of possibleSelectors) {
          const element = document.querySelector(selector);
          if (element && element.scrollHeight > element.clientHeight) {
            scrollableParent = element;
            console.log(`找到可滚动容器: ${selector}`);
            break;
          }
        }
      }
      
      if (scrollableParent) {
        // 滚动到顶部
        scrollableParent.scrollTop = 0;
        console.log('执行滚动到顶部操作');
      }
      
      // 无论是否找到可滚动元素，都尝试滚动页面到顶部
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    }
  }
}
</script>

<style scoped>
.warning-message {
  margin: 20px 0;
}
.analysis-result {
  margin-top: 20px;
  padding: 15px;
  border-radius: 4px;
}
.dark-theme {
  background-color: #1e1e1e;
  color: #ffffff;
}
.result-content {
  white-space: pre-wrap;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  max-height: 500px;
  overflow-y: auto;
  padding: 10px;
}
.model-selection {
  margin: 15px 0;
}
.button-container {
  margin: 20px 0;
}

/* 顶部操作区域的样式 */
.top-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.scroll-top-btn {
  color: #409EFF;
  padding: 5px;
}

.scroll-top-btn:hover {
  color: #66b1ff;
}

.to-top-btn {
  color: #409EFF;
  margin-left: 5px;
  padding: 5px;
}

.to-top-btn:hover {
  color: #66b1ff;
}

/* Markdown样式 - 深色主题 */
.dark-theme >>> h1, 
.dark-theme >>> h2, 
.dark-theme >>> h3, 
.dark-theme >>> h4, 
.dark-theme >>> h5, 
.dark-theme >>> h6 {
  color: #ffffff;
}

.dark-theme >>> p, 
.dark-theme >>> li, 
.dark-theme >>> ul, 
.dark-theme >>> ol {
  color: #e6e6e6;
}

.dark-theme >>> a {
  color: #4da6ff;
}

.dark-theme >>> code {
  background-color: #2d2d2d;
  color: #e6e6e6;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.9em;
}

.dark-theme >>> pre {
  background-color: #2d2d2d;
  padding: 1em;
  border-radius: 5px;
  overflow-x: auto;
  margin: 1em 0;
}

.dark-theme >>> pre code {
  background-color: transparent;
  padding: 0;
  color: #e6e6e6;
}

.dark-theme >>> blockquote {
  border-left: 4px solid #444444;
  padding-left: 1em;
  color: #cccccc;
}

.dark-theme >>> table {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
}

.dark-theme >>> th, .dark-theme >>> td {
  border: 1px solid #444444;
  padding: 8px;
  text-align: left;
}

.dark-theme >>> th {
  background-color: #2d2d2d;
}

.dark-theme >>> tr:nth-child(even) {
  background-color: #2a2a2a;
}

/* Markdown表格样式 */
.dark-theme >>> .markdown-table {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
  background-color: #1e1e1e;
}

.dark-theme >>> .markdown-table th {
  background-color: #2d2d2d;
  color: #ffffff;
  font-weight: bold;
  padding: 8px;
  border: 1px solid #444444;
  text-align: left;
}

.dark-theme >>> .markdown-table td {
  padding: 8px;
  border: 1px solid #444444;
  color: #e6e6e6;
}

.dark-theme >>> .markdown-table tr:nth-child(even) {
  background-color: #2a2a2a;
}

/* 聊天界面样式 */
.chat-interface {
  margin-top: 20px;
}

.chat-messages {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #444;
  border-radius: 4px;
  padding: 10px;
  margin-bottom: 10px;
  background-color: #1e1e1e;
}

/* 更新聊天历史样式 */
.chat-history {
  margin-top: 20px;
  border-top: 1px solid #444;
  padding-top: 15px;
}

.chat-message {
  margin-bottom: 15px;
  padding: 8px 12px;
  border-radius: 8px;
  max-width: 85%;
}

.chat-message.user {
  background-color: #2b5278;
  color: #fff;
  align-self: flex-end;
  margin-left: auto;
}

.chat-message.assistant {
  background-color: #2d2d2d;
  color: #e6e6e6;
  align-self: flex-start;
}

.message-content {
  word-break: break-word;
}

.message-content p:first-child {
  margin-top: 0;
}

.message-content p:last-child {
  margin-bottom: 0;
}

.chat-input {
  display: flex;
  margin-top: 10px;
}

.chat-input .el-textarea {
  flex: 1;
  margin-right: 10px;
}

.chat-input .el-button {
  align-self: flex-end;
}
</style> 