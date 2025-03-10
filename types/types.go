package types

type TranslateRequest struct {
	// Segments       []map[string]interface{} `json:"segments"`
	TargetLanguage string `json:"targetLanguage"`
	ProcessId      string `json:"processId"`
}

type TTSRequest struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

type MediaStorageData struct {
	FileName       string `json:"filename"`
	FileExt        string `json:"fileExt"`
	FileFullName   string `json:"fileFullName"`
	FileUniqueName string `json:"fileUniqueName"`
	FilePath       string `json:"filePath"`
}

type Segment struct {
	Id    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
type TTSSegment struct {
	Id       int     `json:"id"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Text     string  `json:"text"`
	Language string  `json:"language"`
}

type WhisperResponse struct {
	Segments []Segment `json:"segments"`
}

type GetMediaRequest struct {
	FilePath string `json:"filepath"`
}

type ExportVideoRequest struct {
	ProcessId     string                   `json:"processId"`
	Segments      []map[string]interface{} `json:"segments"`
	Language      string                   `json:"language"`
	IsShowCaption bool                     `json:"isShowCaption"`
	IsAppendTTS   bool                     `json:"isAppendTTS"`
	IsTrimVideo   bool                     `json:"isTrimVideo"`
	TrimStart     float64                  `json:"trimStart"`
	TrimEnd       float64                  `json:"trimEnd"`
}
