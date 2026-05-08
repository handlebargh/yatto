// Copyright 2025-2026 handlebargh and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package colors defines functions for all color values
// used by the application.
package colors

import (
	"image/color"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"
)

// Red returns a color value for red.
func Red() color.Color {
	return lipgloss.Color(viper.GetString("colors.red_dark"))
}

// VividRed returns a color value for vivid red.
func VividRed() color.Color {
	return lipgloss.Color(viper.GetString("colors.vividred_dark"))
}

// Indigo returns a color value for indigo.
func Indigo() color.Color {
	return lipgloss.Color(viper.GetString("colors.indigo_dark"))
}

// Green returns a color value for green.
func Green() color.Color {
	return lipgloss.Color(viper.GetString("colors.green_dark"))
}

// Orange returns a color value for orange.
func Orange() color.Color {
	return lipgloss.Color(viper.GetString("colors.orange_dark"))
}

// Blue returns a color value for blue.
func Blue() color.Color {
	return lipgloss.Color(viper.GetString("colors.blue_dark"))
}

// Yellow returns a color value for yellow.
func Yellow() color.Color {
	return lipgloss.Color(viper.GetString("colors.yellow_dark"))
}

// BadgeText returns a color value for badge text.
func BadgeText() color.Color {
	return lipgloss.Color(viper.GetString("colors.badge_text_dark"))
}

// FormTheme returns a huh.Theme based on the configured theme name.
//
// It reads the configuration key "colors.form.theme" using Viper and returns the
// corresponding predefined theme from the huh package. Supported theme values are:
//
//   - "Charm"
//   - "Dracula"
//   - "Catppuccin"
//   - "Base16"
//   - "Base"
//
// If the configuration key is unset or does not match any of the supported values,
// the function defaults to returning ThemeBase16.
// Note: Themes in huh v2 are passed as ThemeFunc, not as pointers.
func FormTheme() huh.Theme {
	switch viper.GetString("colors.form.theme") {
	case "Charm":
		return huh.ThemeFunc(huh.ThemeCharm)
	case "Dracula":
		return huh.ThemeFunc(huh.ThemeDracula)
	case "Catppuccin":
		return huh.ThemeFunc(huh.ThemeCatppuccin)
	case "Base16":
		return huh.ThemeFunc(huh.ThemeBase16)
	case "Base":
		return huh.ThemeFunc(huh.ThemeBase)
	default:
		return huh.ThemeFunc(huh.ThemeBase16)
	}
}
