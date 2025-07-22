package llm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/KubeOperator/kubepi/internal/api/v1/commons"
	v1Llm "github.com/KubeOperator/kubepi/internal/model/v1/llm"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/KubeOperator/kubepi/internal/service/v1/llm"
	pkgV1 "github.com/KubeOperator/kubepi/pkg/api/v1"
	"github.com/asdine/storm/v3"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

type Handler struct {
	llmService llm.Service
}

func NewHandler() *Handler {
	return &Handler{
		llmService: llm.NewService(),
	}
}

func (h *Handler) RegisterRoutes(party iris.Party) {
	party.Post("", h.Create())
	party.Get("", h.List())
	party.Get("/{name}", h.Get())
	party.Put("/{name}", h.Update())
	party.Delete("/{name}", h.Delete())
	party.Post("/{name}/test", h.TestConnection())
	party.Post("/search", h.Search())
}

func (h *Handler) Create() iris.Handler {
	return func(ctx *context.Context) {
		var req v1Llm.LLMModel
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
			return
		}
		if err := h.llmService.Create(&req, common.DBOptions{}); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", req)
	}
}

func (h *Handler) Get() iris.Handler {
	return func(ctx *context.Context) {
		name := ctx.Params().GetString("name")
		model, err := h.llmService.Get(name, common.DBOptions{})
		if err != nil {
			if errors.Is(err, storm.ErrNotFound) {
				ctx.StatusCode(iris.StatusNotFound)
				ctx.Values().Set("message", err.Error())
				return
			}
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", model)
	}
}

func (h *Handler) List() iris.Handler {
	return func(ctx *context.Context) {
		models, err := h.llmService.List(common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", models)
	}
}

func (h *Handler) Update() iris.Handler {
	return func(ctx *context.Context) {
		name := ctx.Params().GetString("name")
		var req v1Llm.LLMModel
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
			return
		}
		req.Name = name
		// 设置元数据的Name字段
		req.Metadata.Name = req.Name
		if err := h.llmService.Update(&req, common.DBOptions{}); err != nil {
			if errors.Is(err, storm.ErrNotFound) {
				ctx.StatusCode(iris.StatusNotFound)
				ctx.Values().Set("message", err.Error())
				return
			}
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", req)
	}
}

func (h *Handler) Delete() iris.Handler {
	return func(ctx *context.Context) {
		name := ctx.Params().GetString("name")
		if err := h.llmService.Delete(name, common.DBOptions{}); err != nil {
			if errors.Is(err, storm.ErrNotFound) {
				ctx.StatusCode(iris.StatusNotFound)
				ctx.Values().Set("message", err.Error())
				return
			}
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", nil)
	}
}

func (h *Handler) TestConnection() iris.Handler {
	return func(ctx *context.Context) {
		name := ctx.Params().GetString("name")
		var req v1Llm.TestRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
			return
		}

		// 检查是否为流式请求
		if req.Stream {
			// 设置响应头为流式传输
			ctx.ResponseWriter().Header().Set("Content-Type", "text/event-stream")
			ctx.ResponseWriter().Header().Set("Cache-Control", "no-cache")
			ctx.ResponseWriter().Header().Set("Connection", "keep-alive")
			ctx.ResponseWriter().Header().Set("Access-Control-Allow-Origin", "*")
			
			// 通过自定义响应写入函数处理流式响应
			streamWriter := func(data []byte) (int, error) {
				// 通过iris.ResponseWriter向客户端写入数据
				n, err := ctx.ResponseWriter().Write(data)
				// 刷新缓冲区确保数据立即发送
				ctx.ResponseWriter().Flush()
				return n, err
			}
			
			// 调用LLM服务处理流式请求
			err := h.llmService.TestConnectionStream(name, &req, streamWriter, common.DBOptions{})
			if err != nil {
				// 错误处理 - 尝试发送错误信息作为SSE事件
				errorMsg := fmt.Sprintf("data: {\"error\":{\"message\":\"%s\"}}\n\n", err.Error())
				ctx.ResponseWriter().Write([]byte(errorMsg))
				ctx.ResponseWriter().Flush()
				return
			}
			
			// 发送结束标志
			ctx.ResponseWriter().Write([]byte("data: [DONE]\n\n"))
			ctx.ResponseWriter().Flush()
			return
		}
		
		// 非流式请求的原有处理
		resp, err := h.llmService.TestConnection(name, &req, common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", resp)
	}
}

func (h *Handler) Search() iris.Handler {
	return func(ctx *context.Context) {
		pageNum, _ := ctx.Values().GetInt(pkgV1.PageNum)
		pageSize, _ := ctx.Values().GetInt(pkgV1.PageSize)

		var searchConditions commons.SearchConditions
		if err := ctx.ReadJSON(&searchConditions); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
			return
		}

		// 获取所有模型
		allModels, err := h.llmService.List(common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}

		// 根据条件过滤模型
		var filteredModels []v1Llm.LLMModel
		if len(searchConditions.Conditions) == 0 {
			// 无条件时返回全部结果
			filteredModels = allModels
		} else {
			for _, model := range allModels {
				match := true
				for _, condition := range searchConditions.Conditions {
					switch condition.Field {
					case "name":
						if condition.Operator == "like" {
							if !strings.Contains(strings.ToLower(model.Name), strings.ToLower(condition.Value)) {
								match = false
							}
						} else if condition.Operator == "eq" {
							if model.Name != condition.Value {
								match = false
							}
						}
					case "baseUri":
						if condition.Operator == "like" {
							if !strings.Contains(strings.ToLower(model.BaseURI), strings.ToLower(condition.Value)) {
								match = false
							}
						} else if condition.Operator == "eq" {
							if model.BaseURI != condition.Value {
								match = false
							}
						}
					case "modelName":
						if condition.Operator == "like" {
							if !strings.Contains(strings.ToLower(model.ModelName), strings.ToLower(condition.Value)) {
								match = false
							}
						} else if condition.Operator == "eq" {
							if model.ModelName != condition.Value {
								match = false
							}
						}
					case "status":
						if condition.Operator == "like" {
							if !strings.Contains(strings.ToLower(model.Status), strings.ToLower(condition.Value)) {
								match = false
							}
						} else if condition.Operator == "eq" {
							if model.Status != condition.Value {
								match = false
							}
						}
					}
				}
				if match {
					filteredModels = append(filteredModels, model)
				}
			}
		}

		// 分页处理
		total := len(filteredModels)
		start := (pageNum - 1) * pageSize
		end := start + pageSize
		if start >= total {
			ctx.Values().Set("data", pkgV1.Page{Items: []v1Llm.LLMModel{}, Total: total})
			return
		}
		if end > total {
			end = total
		}
		ctx.Values().Set("data", pkgV1.Page{Items: filteredModels[start:end], Total: total})
	}
}

func Install(parent iris.Party) {
	handler := NewHandler()
	sp := parent.Party("/llmmodels")
	handler.RegisterRoutes(sp)
} 