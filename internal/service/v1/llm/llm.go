package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/KubeOperator/kubepi/internal/model/v1/llm"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/asdine/storm/v3"
	"github.com/google/uuid"
)

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// WriteFunc 定义一个写入函数类型，用于处理流式响应
type WriteFunc func([]byte) (int, error)

type Service interface {
	common.DBService
	Get(name string, options common.DBOptions) (*llm.LLMModel, error)
	List(options common.DBOptions) ([]llm.LLMModel, error)
	Create(model *llm.LLMModel, options common.DBOptions) error
	Update(model *llm.LLMModel, options common.DBOptions) error
	Delete(name string, options common.DBOptions) error
	TestConnection(name string, testReq *llm.TestRequest, options common.DBOptions) (*llm.ChatResponse, error)
	TestConnectionStream(name string, testReq *llm.TestRequest, writer WriteFunc, options common.DBOptions) error
}

type service struct {
	common.DefaultDBService
}

func NewService() Service {
	return &service{}
}

func (s *service) Get(name string, options common.DBOptions) (*llm.LLMModel, error) {
	db := s.GetDB(options)
	var model llm.LLMModel
	if err := db.One("Name", name, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *service) List(options common.DBOptions) ([]llm.LLMModel, error) {
	db := s.GetDB(options)
	var models []llm.LLMModel
	if err := db.All(&models); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return []llm.LLMModel{}, nil
		}
		return nil, err
	}
	return models, nil
}

func (s *service) Create(model *llm.LLMModel, options common.DBOptions) error {
	db := s.GetDB(options)
	
	// 生成唯一ID
	id := uuid.New().String()
	model.ID = id
	model.Metadata.UUID = id  // 确保设置Metadata.UUID字段，因为storm将使用它作为主键
	
	// 设置其他必要字段
	model.CreatedAt = time.Now()
	model.Status = "untested"
	
	return db.Save(model)
}

func (s *service) Update(model *llm.LLMModel, options common.DBOptions) error {
	db := s.GetDB(options)
	var old llm.LLMModel
	if err := db.One("Name", model.Name, &old); err != nil {
		return err
	}
	
	// 保持ID一致性
	model.ID = old.ID
	model.Metadata.UUID = old.ID  // 确保Metadata.UUID和ID保持一致
	
	// 保留原始创建时间
	model.CreatedAt = old.CreatedAt
	
	return db.Save(model)
}

func (s *service) Delete(name string, options common.DBOptions) error {
	db := s.GetDB(options)
	var model llm.LLMModel
	if err := db.One("Name", name, &model); err != nil {
		return err
	}
	return db.DeleteStruct(&model)
}

func (s *service) TestConnection(name string, testReq *llm.TestRequest, options common.DBOptions) (*llm.ChatResponse, error) {
	model, err := s.Get(name, options)
	if err != nil {
		return nil, err
	}

	// Create chat request
	chatReq := llm.ChatRequest{
		Model: model.ModelName,
		Messages: []llm.ChatMessage{
			{
				Role:    "system",
				Content: "你是一个k8s的专家，请帮我分析pod的日志和event后，给出pod故障的原因以及解决方案.",
			},
			{
				Role:    "user",
				Content: testReq.Content,
			},
		},
		Temperature: model.Temperature,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		fmt.Printf("JSON序列化请求失败: %v\n", err)
		return nil, err
	}

	// 打印完整的请求JSON
	fmt.Printf("发送到LLM的完整请求JSON: %s\n", string(jsonData))

	// Create HTTP request with longer timeout (180秒)
	client := &http.Client{
		Timeout: 180 * time.Second, // 设置180秒超时，给大型模型足够的响应时间
	}
	endpoint := model.BaseURI
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建HTTP请求失败: %v\n", err)
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if model.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", model.APIKey))
	}

	// 记录请求开始时间，用于日志
	startTime := time.Now()
	fmt.Printf("开始请求LLM API: %s, 时间: %s\n", endpoint, startTime.Format("2006-01-02 15:04:05"))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		// 记录请求失败信息
		fmt.Printf("请求LLM API失败: %v, 耗时: %v\n", err, time.Since(startTime))
		// Update model status to unavailable
		model.Status = "unavailable"
		model.LastTestTime = time.Now()
		s.Update(model, options)
		return nil, err
	}
	defer resp.Body.Close()
	
	// 记录请求完成时间
	fmt.Printf("LLM API请求完成, 状态码: %d, 耗时: %v\n", resp.StatusCode, time.Since(startTime))

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return nil, err
	}

	// 打印API响应进行调试
	fmt.Println("API Response:", string(body))

	// 首先尝试解析外层响应
	var outerResp map[string]interface{}
	if err := json.Unmarshal(body, &outerResp); err != nil {
		fmt.Printf("解析外层响应失败: %v\n", err)
		return nil, fmt.Errorf("解析API响应失败: %v", err)
	}
	
	// 检查是否有data字段，这可能是KubePi API的包装
	var innerBody []byte
	if data, ok := outerResp["data"]; ok {
		// 如果data是字符串，尝试解析它
		if dataStr, ok := data.(string); ok {
			innerBody = []byte(dataStr)
		} else {
			// 如果data是对象，将其转换为JSON
			innerBody, err = json.Marshal(data)
			if err != nil {
				fmt.Printf("序列化内层data失败: %v\n", err)
				innerBody = body // 回退到使用原始响应
			}
		}
	} else {
		// 如果没有data字段，使用原始响应
		innerBody = body
	}
	
	fmt.Println("处理后的内层响应:", string(innerBody))
	
	// 尝试解析为标准ChatResponse格式
	var chatResp llm.ChatResponse
	if err := json.Unmarshal(innerBody, &chatResp); err != nil {
		fmt.Printf("解析为ChatResponse失败: %v\n", err)
		
		// 创建一个基本的响应结构
		chatResp = llm.ChatResponse{
			ID:      fmt.Sprintf("response-%d", time.Now().Unix()),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   model.ModelName,
			Choices: []llm.ChatChoice{},
			Usage: llm.ChatUsage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		}
		
		// 尝试从原始响应中提取内容
		content := string(innerBody)
		
		// 处理可能的</think>标记
		thinkEndIndex := strings.Index(content, "</think>")
		if thinkEndIndex != -1 {
			fmt.Printf("检测到</think>标记，位置: %d\n", thinkEndIndex)
			content = content[thinkEndIndex+8:]
			fmt.Printf("提取后的内容长度: %d\n", len(content))
		}
		
		// 添加到响应中
		chatResp.Choices = append(chatResp.Choices, llm.ChatChoice{
			Index: 0,
			Message: llm.ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: "stop",
		})
	} else {
		fmt.Println("成功解析为ChatResponse格式")
		
		// 处理DeepSeek模型的特殊响应格式
		for i := range chatResp.Choices {
			if chatResp.Choices[i].Message.Content != "" {
				// 如果内容中包含</think>标记，只保留后面的内容
				content := chatResp.Choices[i].Message.Content
				thinkEndIndex := strings.Index(content, "</think>")
				if thinkEndIndex != -1 {
					fmt.Printf("检测到</think>标记，位置: %d\n", thinkEndIndex)
					chatResp.Choices[i].Message.Content = content[thinkEndIndex+8:]
					fmt.Printf("提取后的内容长度: %d\n", len(chatResp.Choices[i].Message.Content))
				}
				
				// 检查内容是否是JSON字符串，如果是，尝试解析它
				if strings.HasPrefix(chatResp.Choices[i].Message.Content, "{") && strings.HasSuffix(chatResp.Choices[i].Message.Content, "}") {
					var contentObj map[string]interface{}
					if err := json.Unmarshal([]byte(chatResp.Choices[i].Message.Content), &contentObj); err == nil {
						// 如果是JSON对象，检查是否有detail字段，这可能是错误消息
						if detail, ok := contentObj["detail"].(string); ok && detail == "Not Found" {
							chatResp.Choices[i].Message.Content = "API返回错误: 资源未找到。请检查模型配置和连接。"
						}
					}
				}
			}
		}
	}
	
	// 确保始终有一个有效的响应
	if len(chatResp.Choices) == 0 {
		fmt.Println("响应中没有choices，添加默认choice")
		chatResp.Choices = append(chatResp.Choices, llm.ChatChoice{
			Index: 0,
			Message: llm.ChatMessage{
				Role:    "assistant",
				Content: "无法获取AI回答，请检查模型配置和连接。",
			},
			FinishReason: "stop",
		})
	}
	
	// 打印最终处理后的响应
	finalContent := chatResp.Choices[0].Message.Content
	fmt.Printf("最终处理后的响应内容(前100个字符): %s\n", finalContent[:min(100, len(finalContent))])

	// Update model status
	model.Status = "available"
	model.LastTestTime = time.Now()
	s.Update(model, options)

	return &chatResp, nil
}

// TestConnectionStream 处理流式LLM响应
func (s *service) TestConnectionStream(name string, testReq *llm.TestRequest, writer WriteFunc, options common.DBOptions) error {
	model, err := s.Get(name, options)
	if err != nil {
		return err
	}
	
	// 创建带流式参数的请求
	chatReq := llm.ChatRequest{
		Model: model.ModelName,
		Messages: []llm.ChatMessage{
			{
				Role:    "system",
				Content: "你是一个k8s的专家，请帮我分析pod的日志和event后，给出pod故障的原因以及解决方案.",
			},
			{
				Role:    "user",
				Content: testReq.Content,
			},
		},
		Temperature: model.Temperature,
		Stream:      true, // 启用流式响应
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		fmt.Printf("JSON序列化流式请求失败: %v\n", err)
		return err
	}
	
	// 打印完整的请求JSON
	fmt.Printf("发送到LLM的流式请求JSON: %s\n", string(jsonData))
	
	// 创建带较长超时的HTTP客户端
	client := &http.Client{
		Timeout: 300 * time.Second, // 流式请求使用更长的超时时间
	}
	
	endpoint := model.BaseURI
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建流式HTTP请求失败: %v\n", err)
		return err
	}
	
	// 设置HTTP头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if model.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", model.APIKey))
	}
	
	// 记录请求开始时间
	startTime := time.Now()
	fmt.Printf("开始流式请求LLM API: %s, 时间: %s\n", endpoint, startTime.Format("2006-01-02 15:04:05"))
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		// 记录请求失败信息
		fmt.Printf("流式请求LLM API失败: %v, 耗时: %v\n", err, time.Since(startTime))
		// 更新模型状态为不可用
		model.Status = "unavailable"
		model.LastTestTime = time.Now()
		s.Update(model, options)
		return err
	}
	defer resp.Body.Close()
	
	// 记录请求完成开始接收流式数据的时间
	fmt.Printf("LLM API流式响应开始, 状态码: %d, 等待时间: %v\n", resp.StatusCode, time.Since(startTime))
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试读取错误信息
		body, _ := ioutil.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("LLM API返回非200状态码: %d, 响应: %s", resp.StatusCode, string(body))
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	
	// 创建缓冲读取器
	reader := bufio.NewReader(resp.Body)
	
	// 用于跟踪处理的消息数量
	messageCount := 0
	
	// 持续读取流式响应
	for {
		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 正常结束
				break
			}
			fmt.Printf("读取流式响应行失败: %v\n", err)
			return err
		}
		
		// 跳过空行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 处理SSE格式的行 (data: {...})
		if strings.HasPrefix(line, "data:") {
			// 提取数据部分
			data := strings.TrimSpace(line[5:])
			
			// 检查是否是结束标记
			if data == "[DONE]" {
				fmt.Println("收到流式响应结束标记")
				break
			}
			
			// 直接将SSE格式的行写入响应
			messageData := []byte(line + "\n\n")
			_, err := writer(messageData)
			if err != nil {
				fmt.Printf("写入流式响应失败: %v\n", err)
				return err
			}
			
			messageCount++
			if messageCount % 5 == 0 {
				fmt.Printf("已处理 %d 条流式消息\n", messageCount)
			}
			
			continue
		}
		
		// 非SSE格式的响应行，尝试解析为JSON并转换为SSE格式
		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(line), &chunk); err == nil {
			// 成功解析为JSON，转换为SSE格式
			sseData := fmt.Sprintf("data: %s\n\n", line)
			_, err := writer([]byte(sseData))
			if err != nil {
				fmt.Printf("写入转换后的SSE数据失败: %v\n", err)
				return err
			}
			
			messageCount++
			continue
		}
		
		// 如果不是JSON也不是SSE格式，则直接以SSE格式包装发送
		sseData := fmt.Sprintf("data: %s\n\n", line)
		_, err = writer([]byte(sseData))
		if err != nil {
			fmt.Printf("写入非JSON非SSE行失败: %v\n", err)
			return err
		}
	}
	
	// 记录流式处理完成信息
	fmt.Printf("流式响应处理完成，共处理 %d 条消息, 总耗时: %v\n", messageCount, time.Since(startTime))
	
	// 更新模型状态
	model.Status = "available"
	model.LastTestTime = time.Now()
	s.Update(model, options)
	
	return nil
} 