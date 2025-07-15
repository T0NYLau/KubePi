package common

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Conditions interface{}

// ConditionsMap represents the older map format of conditions
type ConditionsMap map[string]Condition

// ConditionsSlice represents the newer slice format of conditions
type ConditionsSlice []Condition
