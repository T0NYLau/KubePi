package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/KubeOperator/kubepi/internal/model/v1/llm"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/asdine/storm/v3"
	"github.com/google/uuid"
)

type Service interface {
	common.DBService
	Get(name string, options common.DBOptions) (*llm.LLMModel, error)
	List(options common.DBOptions) ([]llm.LLMModel, error)
	Create(model *llm.LLMModel, options common.DBOptions) error
	Update(model *llm.LLMModel, options common.DBOptions) error
	Delete(name string, options common.DBOptions) error
	TestConnection(name string, testReq *llm.TestRequest, options common.DBOptions) (*llm.ChatResponse, error)
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
				Content: "You are a helpful assistant for Kubernetes.",
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
		return nil, err
	}

	// Create HTTP request
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/v1/chat/completions", model.BaseURI)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if model.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", model.APIKey))
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		// Update model status to unavailable
		model.Status = "unavailable"
		model.LastTestTime = time.Now()
		s.Update(model, options)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response
	var chatResp llm.ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	// Update model status
	if chatResp.Error != nil {
		model.Status = "unavailable"
	} else {
		model.Status = "available"
	}
	model.LastTestTime = time.Now()
	s.Update(model, options)

	return &chatResp, nil
} 