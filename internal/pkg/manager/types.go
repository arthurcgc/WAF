package manager

type CreateArgs struct {
	Name      string
	Replicas  int
	Namespace string
	PlanName  string
	ProxyPass string
	Bind      Bind
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
