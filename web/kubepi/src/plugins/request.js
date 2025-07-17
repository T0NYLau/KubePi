import axios from "axios"
import { $error} from "./message"
// import store from "@/store"
import {getLanguage} from "@/i18n"

const instance = axios.create({
    baseURL: "/kubepi", // url = base url + request url
    withCredentials: true,
    timeout: 180000 // request timeout, increase to 3 minutes
})


instance.interceptors.request.use(
    config => {
        config.headers["lang"] = getLanguage()
        return config
    },
    error => {
        return Promise.reject(error)
    }
)

const checkAuth = response => {
    // 请根据实际需求修改
    if (response.status === 401) {
        // let message = i18n.t("commons.login.expires")
        // $alert(message, () => {
        //     store.dispatch("user/logout").then(() => {
        //         location.reload()
        //     })
        // })
    }
}


// 请根据实际需求修改
instance.interceptors.response.use(response => {
    checkAuth(response)
    
    // 特殊处理LLM API响应
    if (response.config.url.includes('/llmmodels') && response.config.url.includes('/test')) {
        console.log('检测到LLM测试请求响应');
        
        // 如果响应是字符串，尝试解析为JSON
        if (typeof response.data === 'string') {
            try {
                console.log('LLM响应是字符串，尝试解析为JSON');
                response.data = JSON.parse(response.data);
            } catch (e) {
                console.log('无法将LLM响应解析为JSON，保持原样');
            }
        }
        
        // 处理DeepSeek特殊格式
        if (response.data && 
            response.data.choices && 
            response.data.choices.length > 0 && 
            response.data.choices[0].message && 
            response.data.choices[0].message.content) {
            
            const content = response.data.choices[0].message.content;
            const thinkEndIndex = content.indexOf('</think>');
            if (thinkEndIndex !== -1) {
                console.log('在全局拦截器中检测到</think>标记');
                response.data.choices[0].message.content = content.substring(thinkEndIndex + 8).trim();
            }
        }
    }
    
    return response
}, error => {
    let msg
    if (error.response) {
        checkAuth(error.response)
        msg = error.response.data.message || error.response.data
    } else {
        msg = error.message
    }
    $error(msg)
    return Promise.reject(error)
})

export const request = instance

/* 简化请求方法，统一处理返回结果，并增加loading处理，这里以{success,data,message}格式的返回值为例，具体项目根据实际需求修改 */
const promise = (request, loading = {}) => {
    return new Promise((resolve, reject) => {
        loading.status = true
        request.then(response => {
            if (response.data.success || (response.status==200 && response.data.kind=="SecretList")) {
                resolve(response.data)
            } else {
                reject(response.message)
            }
            loading.status = false
        }).catch(error => {
            reject(error)
            loading.status = false
        })
    })
}

export const get = (url, data, loading) => {
    return promise(request({url: url, method: "get", params: data}), loading)
}

export const post = (url, data, loading) => {
    return promise(request({url: url, method: "post", data}), loading)
}

export const put = (url, data, loading) => {
    return promise(request({url: url, method: "put", data}), loading)
}

export const del = (url, loading) => {
    return promise(request({url: url, method: "delete"}), loading)
}

export const patch = (url, data, headers, loading) => {
    if (headers) {
        return promise(request({url: url, headers: headers, method: "patch", data}), loading)
    }
    return promise(request({url: url, method: "patch", data}), loading)
}

export default {
    install(Vue) {
        Vue.prototype.$get = get
        Vue.prototype.$post = post
        Vue.prototype.$put = put
        Vue.prototype.$delete = del
        Vue.prototype.$request = request
    }
}
