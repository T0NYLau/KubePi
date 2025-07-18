import {get, post, del, put} from "@/plugins/request"
import axios from "axios"
import { $error } from "@/plugins/message"

const baseUrl = "/api/v1/llmmodels"

export function getLLMModels() {
    return get(baseUrl)
}

export function getLLMModel(name) {
    return get(`${baseUrl}/${name}`)
}

export function createLLMModel(model) {
    return post(baseUrl, model)
}

export function updateLLMModel(name, model) {
    return put(`${baseUrl}/${name}`, model)
}

export function deleteLLMModel(name) {
    return del(`${baseUrl}/${name}`)
}

export function testLLMModel(name, content, options = {}) {
    // 使用自定义axios实例，单独为LLM请求设置更长的超时时间
    const llmAxios = axios.create({
        baseURL: "/kubepi",
        withCredentials: true,
        timeout: 300000, // 5分钟超时，给大型模型留出足够的响应时间
    });
    
    // 记录请求开始时间
    console.log(`开始LLM测试请求: ${new Date().toISOString()}, 模型名称: ${name}, 内容长度: ${content.length}, 流式模式: ${options.streaming ? '是' : '否'}`);

    // 如果是流式传输模式
    if (options.streaming) {
        const requestOptions = {
            method: 'POST',
            url: `${baseUrl}/${name}/test`,
            data: {content, stream: true},
            responseType: 'text',
        };
        
        // 添加进度回调
        if (options.onProgress && typeof options.onProgress === 'function') {
            requestOptions.onDownloadProgress = function(progressEvent) {
                // 确保我们有响应数据
                if (progressEvent.currentTarget && typeof progressEvent.currentTarget.response === 'string') {
                    const responseText = progressEvent.currentTarget.response;
                    options.onProgress(progressEvent);
                    
                    // 调试信息 - 每次进度更新时记录接收到的新数据量
                    console.log(`流数据更新: 收到 ${responseText.length} 字节的数据，当前时间: ${new Date().toISOString()}`);
                }
            };
        }

        // 如果提供了AbortController，添加到请求选项
        if (options.signal) {
            requestOptions.signal = options.signal;
        }

        // 确保发送正确的流式请求标志
        requestOptions.headers = {
            'Content-Type': 'application/json',
            'Accept': 'text/event-stream'
        };
        
        return llmAxios(requestOptions);
    }
    
    // 非流式模式 - 使用原有实现
    return new Promise((resolve, reject) => {
        llmAxios.post(`${baseUrl}/${name}/test`, {content})
            .then(response => {
                // 记录请求完成时间和状态
                console.log(`LLM测试请求完成: ${new Date().toISOString()}, 状态: ${response.status}`);
                resolve(response);
            })
            .catch(error => {
                console.error('LLM测试请求失败:', error);
                $error(error.response?.data?.message || error.message || '请求失败');
                reject(error);
            });
    });
}

export function searchLLMModels(pageNum, pageSize, conditions) {
    let url = `${baseUrl}/search?pageNum=${pageNum}&&pageSize=${pageSize}`
    return post(url, {conditions: conditions})
} 