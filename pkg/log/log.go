package log

type Log struct {
	Path               string `json:"path"`
	RequestBody        string `json:"request_body"`
	ResponseBody       string `json:"response_body"`
	Service            string `json:"service"`
	ResponseStatusCode int    `json:"response_status_code"`
	RequestHeaders     string `json:"request_headers"`
	ResponseHeaders    string `json:"response_headers"`
}
