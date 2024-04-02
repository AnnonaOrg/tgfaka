package constvar

func APPName() string {
	return APP_NAME
}
func APPVersion() string {
	return "v" + APP_VERSION
}

func APPAbout() string {
	text := APPName() + " " + APPVersion()
	return text
}
