package backboard

func CheckTBL(key int, guiNum int, eye byte, chi bool) bool {

	return GetMjDataTableInstance().Check(key, guiNum, eye, chi)
}
