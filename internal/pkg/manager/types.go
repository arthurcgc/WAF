package manager

type CreateArgs struct {
	Name      string
	Replicas  int
	Namespace string
	PlanName  string
	Bind      Bind
	Rules     Rules
}

type UpdateArgs struct {
	Name      string
	Replicas  int
	Namespace string
	PlanName  string
	Bind      Bind
	Rules     Rules
}

type Rules struct {
	CustomRules           []string `json:"customRules,omitempty"`
	EnableDefaultHoneyPot bool     `json:"defaultHoney,omitempty"`
}

type Bind struct {
	ServiceName string
	Namespace   string
	Protocol    string
}

type DeleteArgs struct {
	Name      string
	Namespace string
}
