package commons

import "github.com/KubeOperator/kubepi/internal/service/v1/common"

type SearchConditions struct {
	Conditions []common.Condition `json:"conditions"`
}
