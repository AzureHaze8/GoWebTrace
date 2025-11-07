package main

import (
	"GoWebTrace/cmd"
	"fmt"

	"github.com/fatih/color"
)

func main() {
	banner := `
  _____   _      __     __ ______                
 / ___/__| | /| / /__  / //_  __/_______________
/ (_ / _ \ |/ |/ / -_)/ _\/ /  / __/ _ // _ / -_)
\___/\___/__/|__/\___/.__/_/  /_/\_\__/\\___\___/                                               
`
	const author = "青烟、okaeri"
	const version = "v1.0"

	// 右对齐显示
	infoPadding := 40
	versionLine := fmt.Sprintf("Version: %s", version)
	authorLine := fmt.Sprintf("Author: %s", author)

	fmt.Println(color.CyanString(banner))
	fmt.Printf("%*s\n", infoPadding, color.YellowString(authorLine))
	fmt.Printf("%*s\n\n", infoPadding, color.YellowString(versionLine))


	cmd.Execute()
}
