package main

func mainRequest(r Kitreq) (int, string) {
	codeKit, data := getDataKitVending(r.CompanyId, r.UserLogin, r.PasswordKit, r.Date)
	if codeKit != 0 {
		return codeKit, data
	} else {
		codeFns := check()
		if codeFns == "authentication.failed.expired.token" {
			codeKit = 1
			getToken(r.Device, r.INN, r.PasswordFns)
		}
		return codeKit, codeFns
	}
}
