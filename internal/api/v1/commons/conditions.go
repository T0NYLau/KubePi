package commons

import (
	"encoding/json"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
)

type SearchConditions struct {
	Conditions interface{} `json:"conditions"`
}

// UnmarshalJSON handles both object and array formats for conditions
func (s *SearchConditions) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first (new format)
	var arrayFormat struct {
		Conditions []common.Condition `json:"conditions"`
	}
	if err := json.Unmarshal(data, &arrayFormat); err == nil {
		s.Conditions = arrayFormat.Conditions
		return nil
	}

	// Fall back to object format (old format)
	var objectFormat struct {
		Conditions map[string]common.Condition `json:"conditions"`
	}
	if err := json.Unmarshal(data, &objectFormat); err == nil {
		s.Conditions = objectFormat.Conditions
		return nil
	}

	// If both fail, return the original error
	return json.Unmarshal(data, &struct{
		Conditions interface{} `json:"conditions"`
	}{}) 
}
