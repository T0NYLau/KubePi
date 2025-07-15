import {get, post, del, put} from "@/plugins/request"

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
    return post(`${baseUrl}/${name}/test`, {content})
}

export function searchLLMModels(pageNum, pageSize, conditions) {
    let url = `${baseUrl}/search?pageNum=${pageNum}&&pageSize=${pageSize}`
    return post(url, {conditions: conditions})
} 