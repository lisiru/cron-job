package params



type AddOrUpdateJobOptions struct {
	JobId        string `json:"job_id"`
	DelaySeconds uint64 `json:"delay_seconds"`
	TtrSeconds   uint64 `json:"ttr_seconds"`
	Body         string `json:"body"`
	IsLoop       bool   `json:"is_loop"`
	NotifyUrl    string `json:"notify_url"`
	Stat         int   `json:"stat"`
}
type CommonOptions struct {
	JobId        string `json:"job_id"`
	DelaySeconds uint64 `json:"delay_seconds"`
	TtrSeconds   uint64 `json:"ttr_seconds"`
	Body         string `json:"body"`
	IsLoop       bool   `json:"is_loop"`
}

type DelJobOptions struct {
	JobId string `json:"job_id"`
}

type UpdateJobOptions struct {
	JobId        string `json:"job_id"`
	DelaySeconds uint64 `json:"delay_seconds"`
	TtrSeconds   uint64 `json:"ttr_seconds"`
	Body         string `json:"body"`
	IsLoop       bool   `json:"is_loop"`
	NotifyUrl    string `json:"notify_url"`
	Stat         uint   `json:"stat" defalut:"1"`
}
