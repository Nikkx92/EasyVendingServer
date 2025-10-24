package main

func mainRequest(r Request) (int, string) {
	if r.Data.RefreshToken == "" {
		//getRefreshToken(r.Device, r.Data.INN, r.Data.PasswordFns)

	}
	codeKit, data := getDataKitVending(r.Data.CompanyId, r.Data.UserLogin, r.Data.PasswordKit, r.Data.Date)
	if codeKit != 0 {
		return codeKit, data
	} else {
		codeFns := check()
		if codeFns == "authentication.failed.expired.token" {
			codeKit = 1
			//getToken(r.Device, r.INN, r.PasswordFns)
		}
		return codeKit, codeFns
	}
}
