package kitvending

type auth struct {
	CompanyId string `json:"companyid"`
	RequestId string `json:"requestid"`
	UserLogin string `json:"userlogin"`
	Sign      string `json:"sign"`
}

type filter struct {
	UpDate string `json:"update"`
	ToDate string `json:"todate"`
}

type request struct {
	Auth   auth   `json:"Auth"`
	Filter filter `json:"Filter"`
}

type response struct {
	ErrorMessage string `json:"ErrorMessage"`
	ResultCode   int    `json:"ResultCode"`
	Sales        []sale `json:"Sales"`
}

type sale struct {
	GoodsName string  `json:"GoodsName"`
	Sum       float64 `json:"Sum"`
}
