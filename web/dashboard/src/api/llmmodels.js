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
    console.log(`开始LLM测试请求: ${new Date().toISOString()}, 模型名称: ${name}, 流式模式: ${options.streaming ? '是' : '否'}`);

    // 如果是流式传输模式，使用fetch和ReadableStream
    if (options.streaming) {
        return new Promise((resolve, reject) => {
            // 创建请求配置
            const fetchOptions = {
            method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'text/event-stream'
                },
                body: JSON.stringify({ content, stream: true }),
                credentials: 'include'
        };
        
            // 如果提供了AbortController，添加到请求选项
            if (options.signal) {
                fetchOptions.signal = options.signal;
            }

            // 使用fetch API发送请求
            fetch(`/kubepi${baseUrl}/${name}/test`, fetchOptions)
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! Status: ${response.status}`);
                    }
                    
                    // 获取响应的可读流
                    const reader = response.body.getReader();
                    const decoder = new TextDecoder();
                    let buffer = '';
                    
                    // 使用async函数处理流
                    async function readStream() {
                        try {
                            while (true) {
                                // 读取数据块
                                const { value, done } = await reader.read();
                    
                                // 如果流结束，退出循环
                                if (done) {
                                    console.log('流结束');
                                    break;
                                }
                                
                                // 解码数据块并添加到缓冲区
                                const text = decoder.decode(value, { stream: true });
                                buffer += text;
                                
                                // 按行处理数据
                                const lines = buffer.split('\n');
                                
                                // 保留最后一行（可能不完整）作为新的缓冲区
                                buffer = lines.pop() || '';
                                
                                // 处理每一行
                                for (const line of lines) {
                                    if (line.trim() === '') continue;
                                    
                                    // 处理SSE格式数据
                                    if (line.startsWith('data:')) {
                                        const eventData = line.substring(5).trim();
                                        
                                        // 检查是否是结束标记
                                        if (eventData === '[DONE]') {
                                            console.log('收到[DONE]标记');
                                            continue;
        }

                                        // 将数据传递给回调函数
                                        if (options.onProgress) {
                                            try {
                                                options.onProgress(eventData);
                                            } catch (e) {
                                                console.error('处理进度回调时出错:', e);
                                            }
                                        }
                                    }
                                }
    }
    
                            // 处理缓冲区中剩余的数据
                            if (buffer.trim() !== '' && options.onProgress) {
                                options.onProgress(buffer);
                            }
                            
                            // 解析为空对象的响应，因为实际内容已经通过onProgress回调处理了
                            resolve({ data: {} });
                        } catch (error) {
                            console.error('读取流时出错:', error);
                            reject(error);
                        }
                    }
                    
                    // 开始读取流
                    readStream();
            })
            .catch(error => {
                console.error('LLM测试请求失败:', error);
                    $error(error.message || '请求失败');
                reject(error);
            });
    });
    }
    
    // 非流式模式 - 使用常规请求
    const llmAxios = axios.create({
        baseURL: "/kubepi",
        withCredentials: true,
        timeout: 300000, // 5分钟超时
    });
    
    return llmAxios.post(`${baseUrl}/${name}/test`, { content })
        .catch(error => {
            console.error('LLM测试请求失败:', error);
            $error(error.response?.data?.message || error.message || '请求失败');
            throw error;
        });
}

export function searchLLMModels(pageNum, pageSize, conditions) {
    let url = `${baseUrl}/search?pageNum=${pageNum}&&pageSize=${pageSize}`
    return post(url, {conditions: conditions})
} 