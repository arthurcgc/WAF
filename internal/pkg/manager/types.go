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
	*RulesAfter           `json:"removeAfter,omitempty"`
	CustomRules           []string `json:"customRules,omitempty"`
	EnableDefaultHoneyPot bool     `json:"defaultHoney,omitempty"`
}

type RulesAfter struct {
	// Example Exclusion Rule: Remove a group of rules
	// ModSecurity Rule Exclusion: Disable PHP injection rules
	// SecRuleRemoveByTag "attack-injection-php"
	RemoveByTag []string `json:"removeByTag,omitempty"`
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
