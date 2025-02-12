package types

type TranslateRequest struct {
	Segments       []map[string]interface{} `json:"segments"`
	TargetLanguage string                   `json:"targetLanguage"`
}

type Segment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type WhisperResponse struct {
	Segments []Segment `json:"segments"`
}
