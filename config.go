/*
This file is:

The MIT License (MIT)

# Copyright (c) 2014 Bitrise

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"github.com/bitrise-io/go-steputils/stepconf"
	"os"
)

// Buildstatus
var success = os.Getenv("BITRISE_BUILD_STATUS") == "0"

// Config ...
type Config struct {
	// Settings
	Debug      bool            `env:"is_debug_mode,opt[yes,no]"`
	WebhookURL stepconf.Secret `env:"webhook_url"`
	// Message Main
	CardStyle           string `env:"card_style"`
	CardStyleOnError    string `env:"card_style_on_error"`
	CardHeadline        string `env:"card_headline"`
	CardHeadlineOnError string `env:"card_headline_on_error"`

	Title        string `env:"title"`
	TitleOnError string `env:"title_on_error"`
	// Message Git
	AuthorName string `env:"author_name"`
	Subject    string `env:"subject"`
	// Message Content
	Fields         string `env:"fields"`
	Images         string `env:"images"`
	ImagesOnError  string `env:"images_on_error"`
	Buttons        string `env:"buttons"`
	ButtonsOnError string `env:"buttons_on_error"`
}
