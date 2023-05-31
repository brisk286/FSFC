package pkg

type AllSaveSpaceResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Dirs []string `json:"dirs"`
}
