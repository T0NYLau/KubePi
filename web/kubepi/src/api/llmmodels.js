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

export function testLLMModel(name, content) {
    // 使用自定义axios实例，单独为LLM请求设置更长的超时时间
    const llmAxios = axios.create({
        baseURL: "/kubepi",
        withCredentials: true,
        timeout: 300000, // 5分钟超时，给大型模型留出足够的响应时间
        // 添加响应拦截器，处理特殊格式
        transformResponse: [function(data) {
            // 尝试解析JSON
            try {
                console.log('原始响应数据:', data);
                const jsonData = JSON.parse(data);
                console.log('成功解析响应为JSON');
                return jsonData;
            } catch (e) {
                console.log('响应不是有效的JSON，返回原始数据');
                // 如果不是有效的JSON，返回原始数据
                return data;
            }
        }]
    });
    
    // 记录请求开始时间
    console.log(`开始LLM测试请求: ${new Date().toISOString()}, 模型名称: ${name}, 内容长度: ${content.length}`);
    
    // 返回Promise对象
    return new Promise((resolve, reject) => {
        llmAxios.post(`${baseUrl}/${name}/test`, {content})
            .then(response => {
                // 记录请求完成时间和状态
                console.log(`LLM测试请求完成: ${new Date().toISOString()}, 状态: ${response.status}`);
                console.log('响应数据类型:', typeof response.data);
                
                // 处理各种可能的响应格式
                if (typeof response.data === 'string') {
                    console.log('响应是字符串，尝试解析为JSON');
                    try {
                        // 如果响应是字符串，尝试解析为JSON
                        const jsonData = JSON.parse(response.data);
                        response.data = jsonData;
                        console.log('成功将字符串响应解析为JSON');
                    } catch (e) {
                        console.log('无法将字符串响应解析为JSON，创建兼容结构');
                        // 如果无法解析，创建一个兼容的结构
                        response.data = {
                            choices: [{
                                message: {
                                    role: 'assistant',
                                    content: response.data
                                }
                            }],
                            usage: { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 }
                        };
                    }
                }
                
                // 检查是否有嵌套的JSON字符串
                if (response.data && response.data.choices && 
                    response.data.choices.length > 0 && 
                    response.data.choices[0].message && 
                    response.data.choices[0].message.content) {
                    
                    const content = response.data.choices[0].message.content;
                    
                    // 检查内容是否是JSON字符串
                    if (typeof content === 'string' && 
                        content.trim().startsWith('{') && 
                        content.trim().endsWith('}')) {
                        
                        try {
                            console.log('检测到content可能是JSON字符串，尝试解析');
                            const contentObj = JSON.parse(content);
                            
                            // 检查是否有detail字段，这可能是错误消息
                            if (contentObj.detail === 'Not Found') {
                                response.data.choices[0].message.content = "API返回错误: 资源未找到。请检查模型配置和连接。";
                            } 
                            // 检查是否有data字段，这可能是嵌套的响应
                            else if (contentObj.data) {
                                console.log('从content中提取data字段');
                                
                                // 如果data包含完整的响应结构，使用它替换外层响应
                                if (contentObj.data.choices && contentObj.data.choices.length > 0) {
                                    console.log('从content.data中提取完整响应结构');
                                    response.data = contentObj.data;
                                }
                            }
                        } catch (e) {
                            console.log('解析content为JSON失败:', e);
                        }
                    }
                }
                
                // 确保响应有有效的结构
                if (!response.data) {
                    console.log('响应数据为空，创建默认结构');
                    response.data = {
                        choices: [{
                            message: {
                                role: 'assistant',
                                content: '收到了空响应，请检查模型配置和连接'
                            }
                        }],
                        usage: { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 }
                    };
                } else if (!response.data.choices || response.data.choices.length === 0) {
                    console.log('响应中没有choices数组，创建默认结构');
                    
                    // 尝试从响应中提取内容
                    let content = '';
                    if (typeof response.data === 'string') {
                        content = response.data;
                    } else if (response.data.assistant) {
                        content = response.data.assistant;
                    } else if (response.data.text) {
                        content = response.data.text;
                    } else if (response.data.content) {
                        content = response.data.content;
                    } else {
                        try {
                            content = JSON.stringify(response.data);
                        } catch (e) {
                            content = '无法解析响应内容';
                        }
                    }
                    
                    // 处理可能的</think>标记
                    const thinkEndIndex = content.indexOf('</think>');
                    if (thinkEndIndex !== -1) {
                        content = content.substring(thinkEndIndex + 8).trim();
                    }
                    
                    response.data = {
                        choices: [{
                            message: {
                                role: 'assistant',
                                content: content
                            }
                        }],
                        usage: response.data.usage || { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 }
                    };
                } else if (response.data.choices[0] && !response.data.choices[0].message) {
                    console.log('choices[0]中没有message对象，创建默认message');
                    
                    // 尝试从choice中提取内容
                    let content = '';
                    const choice = response.data.choices[0];
                    if (typeof choice === 'string') {
                        content = choice;
                    } else if (choice.text) {
                        content = choice.text;
                    } else if (choice.content) {
                        content = choice.content;
                    } else {
                        try {
                            content = JSON.stringify(choice);
                        } catch (e) {
                            content = '无法解析choice内容';
                        }
                    }
                    
                    response.data.choices[0].message = {
                        role: 'assistant',
                        content: content
                    };
                }
                
                // 记录响应结构
                if (response.data && typeof response.data === 'object') {
                    console.log('响应数据结构:', Object.keys(response.data).join(', '));
                    if (response.data.choices && response.data.choices.length > 0) {
                        console.log('第一个choice结构:', Object.keys(response.data.choices[0]).join(', '));
                        if (response.data.choices[0].message) {
                            console.log('message结构:', Object.keys(response.data.choices[0].message).join(', '));
                            
                            // 处理DeepSeek特殊格式
                            if (response.data.choices[0].message.content) {
                                const content = response.data.choices[0].message.content;
                                const thinkEndIndex = content.indexOf('</think>');
                                if (thinkEndIndex !== -1) {
                                    console.log('在API层检测到</think>标记，位置:', thinkEndIndex);
                                    response.data.choices[0].message.content = content.substring(thinkEndIndex + 8).trim();
                                    console.log('API层处理后的内容长度:', response.data.choices[0].message.content.length);
                                }
                            }
                        }
                    }
                }
                
                resolve(response);
            })
            .catch(error => {
                console.error('LLM测试请求失败:', error);
                // 提供更详细的错误信息
                if (error.code === 'ECONNABORTED') {
                    console.error('请求超时，请检查模型服务是否正常运行');
                    $error('请求超时，大型模型可能需要更长的响应时间，请检查模型服务是否正常运行');
                } else if (error.response && error.response.status === 404) {
                    console.error('API端点未找到，请检查BaseURI配置');
                    $error('API端点未找到(404)，请检查BaseURI配置是否正确，确保包含完整路径，例如: http://ip:port/v1/chat/completions');
                } else if (error.message && error.message.includes('Network Error')) {
                    console.error('网络错误，无法连接到API服务');
                    $error('网络错误，无法连接到API服务，请检查服务地址是否正确且可访问');
                } else {
                    $error(error.response?.data?.message || error.message || '请求失败');
                }
                reject(error);
            });
    });
}

export function searchLLMModels(pageNum, pageSize, conditions) {
    let url = `${baseUrl}/search?pageNum=${pageNum}&&pageSize=${pageSize}`
    return post(url, {conditions: conditions})
} 