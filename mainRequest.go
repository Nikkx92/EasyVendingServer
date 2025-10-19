package main

func mainRequest(companyId, userLogin, password, date string) (int, string) {
	codeKit, data := getDataKitVending(companyId, userLogin, password, date)
	if codeKit != 0 {
		return codeKit, data
	} else {
		codeFns := check()
		if codeFns == "authentication.failed.expired.token" {
			codeKit = 1
		}
		return codeKit, codeFns
	}
}
