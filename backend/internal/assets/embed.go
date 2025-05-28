package assets

import (
	"embed"
)

//go:embed schemas/*
var Schemas embed.FS

//go:embed ui_schemas/*
var UI_Schemas embed.FS
