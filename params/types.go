package params



type AddOrUpdateJobOptions struct {
	JobId        string `json:"job_id"`
	DelaySeconds int64 `json:"delay_seconds"`
	TtrSeconds   int64 `json:"ttr_seconds"`
	Body         string `json:"body"`
	IsLoop       bool   `json:"is_loop"`
	NotifyUrl    string `json:"notify_url"`
	Stat         int   `json:"stat"`
}


type DelJobOptions struct {
	JobId string `json:"job_id"`
}

type FinishOptions struct {
	JobId string `json:"job_id"`
}


