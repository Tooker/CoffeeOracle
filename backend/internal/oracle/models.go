package oracle

// OracleRequest represents the payload required to generate a coffee fortune.
// Names and textual inputs must be sanitized before being used in prompts.
type OracleRequest struct {
	Name         string `json:"name"`
	Creativity   int    `json:"creativity"`
	QuestionMode bool   `json:"questionMode"`
	Question     string `json:"question"`
	ImageName    string `json:"imageName"`
	ImageMIME    string `json:"imageMime"`
	ImageBase64  string `json:"imageBase64"`
}
