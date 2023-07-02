package qqrobot

import "github.com/gookit/color"

func bold(c color.Color) color.Style {
	return color.Style{color.Bold, c}
}
