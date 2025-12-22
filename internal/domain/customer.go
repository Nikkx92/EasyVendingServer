package domain

type MetaDetails struct {
	UserAgent string `json:"userAgent"`
}

type DeviceInfo struct {
	SourceDeviceID string       `json:"sourceDeviceId"`
	SourceType     string       `json:"sourceType"`
	AppVersion     string       `json:"appVersion"`
	MetaDetails    *MetaDetails `json:"metaDetails"`
}

type LoginData struct {
	CompanyId   string `json:"CompanyId"`
	UserLogin   string `json:"UserLogin"`
	PasswordKit string `json:"PasswordKit"`
	INN         string `json:"INN"`
	PasswordFns string `json:"PasswordFns"`
}

type Customer struct {
	CompanyID    string
	UserLogin    string
	PasswordKit  string
	INN          string
	PasswordFns  string
	DeviceInfo   *DeviceInfo
	RefreshToken string
	Token        string
}
