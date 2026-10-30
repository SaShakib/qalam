package main

import _ "embed"

//go:embed assets/app.html
var appHTML []byte

//go:embed assets/setup.html
var setupHTML []byte

//go:embed assets/guide.html
var guideHTML []byte
