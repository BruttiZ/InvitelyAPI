package uploads

type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}
