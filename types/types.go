package types

type TranslateRequest struct {
	Segments       []map[string]interface{} `json:"segments"`
	TargetLanguage string                   `json:"targetLanguage"`
}

type TTSRequest struct {
	TTSSegments []map[string]interface{} `json:"segments"`
	ProcessID   string                   `json:"processId"`
}

type Segment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
type TTSSegment struct {
	ID       int     `json:"id"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Text     string  `json:"text"`
	Language string  `json:"language"`
}

type WhisperResponse struct {
	Segments []Segment `json:"segments"`
}

type ExportVideoRequest struct {
	ProcessId string                   `json:"processId"`
	Segments  []map[string]interface{} `json:"segments"`
}
