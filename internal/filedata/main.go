package filedata

type FileData struct {
	Data   []byte `json:"data"`
	Cipher string `json:"cipher"`
}
