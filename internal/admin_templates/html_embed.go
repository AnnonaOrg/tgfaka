package admin_templates

import (
	"embed"
)

//go:embed *.html
var TPL embed.FS

//go:embed *.html
//go:embed sdk/*
var StaticSdk embed.FS
