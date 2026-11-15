package support

type Case struct {
	ID       string `json:"id"`
	OrderID  string `json:"order_id"`
	ItemID   string `json:"item_id"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
	Reason   string `json:"reason"`
	Deadline string `json:"deadline"`
	Version  int    `json:"version"`
}
type Assignment struct {
	ID              string `json:"id"`
	CaseID          string `json:"case_id"`
	StoreID         string `json:"store_id"`
	ItemID          string `json:"item_id"`
	Status          string `json:"status"`
	ContractVersion string `json:"contract_version"`
}
