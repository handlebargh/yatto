// Copyright 2025 handlebargh and contributors
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

// package colors defines functions for all color values
// used by the application.
package colors

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// Red returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.red_light" for the light theme
//   - "colors.red_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Red() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.red_light"),
		Dark:  viper.GetString("colors.red_dark"),
	}
}

// VividRed returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.vividRed_light" for the light theme
//   - "colors.vividRed_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func VividRed() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.vividRed_light"),
		Dark:  viper.GetString("colors.vividRed_dark"),
	}
}

// Indigo returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.indigo_light" for the light theme
//   - "colors.indigo_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Indigo() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.indigo_light"),
		Dark:  viper.GetString("colors.indigo_dark"),
	}
}

// Green returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.green_light" for the light theme
//   - "colors.green_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Green() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.green_light"),
		Dark:  viper.GetString("colors.green_dark"),
	}
}

// Orange returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.orange_light" for the light theme
//   - "colors.orange_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Orange() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.orange_light"),
		Dark:  viper.GetString("colors.orange_dark"),
	}
}

// Blue returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.blue_light" for the light theme
//   - "colors.blue_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Blue() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.blue_light"),
		Dark:  viper.GetString("colors.blue_dark"),
	}
}

// Yellow returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.yellow_light" for the light theme
//   - "colors.yellow_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func Yellow() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.yellow_light"),
		Dark:  viper.GetString("colors.yellow_dark"),
	}
}

// BadgeText returns an AdaptiveColor configured for light and dark themes.
//
// The color values are loaded from Viper configuration keys:
//   - "colors.badge_text_light" for the light theme
//   - "colors.Badge_text_dark" for the dark theme
//
// These should be set via Viper's defaults or loaded from a config file
// before calling this function. If not set, the returned color will use empty strings.
func BadgeText() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: viper.GetString("colors.badge_text_light"),
		Dark:  viper.GetString("colors.Badge_text_dark"),
	}
}

// FormTheme returns a pointer to a huh.Theme based on the configured theme name.
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
func FormTheme() *huh.Theme {
	switch viper.GetString("colors.form.theme") {
	case "Charm":
		return huh.ThemeCharm()
	case "Dracula":
		return huh.ThemeDracula()
	case "Catppuccin":
		return huh.ThemeCatppuccin()
	case "Base16":
		return huh.ThemeBase16()
	case "Base":
		return huh.ThemeBase()
	default:
		return huh.ThemeBase16()
	}
}
