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
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"
)

// Red returns a color.Color representing the red color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the red color, modified
//     for the current terminal background.
func Red() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.red_light")),
		lipgloss.Color(viper.GetString("colors.red_dark")),
	)
}

// VividRed returns a color.Color representing the vividred color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the vividred color, modified
//     for the current terminal background.
func VividRed() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.vividred_light")),
		lipgloss.Color(viper.GetString("colors.vividred_dark")),
	)
}

// Indigo returns a color.Color representing the indigo color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the indigo color, modified
//     for the current terminal background.
func Indigo() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.indigo_light")),
		lipgloss.Color(viper.GetString("colors.indigo_dark")),
	)
}

// Green returns a color.Color representing the green color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the green color, modified
//     for the current terminal background.
func Green() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.green_light")),
		lipgloss.Color(viper.GetString("colors.green_dark")),
	)
}

// Orange returns a color.Color representing the orange color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the orange color, modified
//     for the current terminal background.
func Orange() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.orange_light")),
		lipgloss.Color(viper.GetString("colors.orange_dark")),
	)
}

// Blue returns a color.Color representing the blue color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the blue color, modified
//     for the current terminal background.
func Blue() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.blue_light")),
		lipgloss.Color(viper.GetString("colors.blue_dark")),
	)
}

// Yellow returns a color.Color representing the yellow color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the yellow color, modified
//     for the current terminal background.
func Yellow() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.yellow_light")),
		lipgloss.Color(viper.GetString("colors.yellow_dark")),
	)
}

// BadgeText returns a color.Color representing the badge_text color
// as defined in the configuration.
//
// Returns:
//   - A color.Color representing the badge_text color, modified
//     for the current terminal background.
func BadgeText() color.Color {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return lightDark(
		lipgloss.Color(viper.GetString("colors.badge_text_light")),
		lipgloss.Color(viper.GetString("colors.badge_text_dark")),
	)
}

// FormTheme returns a huh.ThemeFunc based on the configured theme name.
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
func FormTheme() huh.ThemeFunc {
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
