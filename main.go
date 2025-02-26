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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/log"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
)

type TeamsMessage struct {
	Type        string          `json:"type"`
	Attachments []Attachment    `json:"attachments"`
}

type Attachment struct {
	ContentType string            `json:"contentType"`
	Content     adaptivecard.Card `json:"content"`
}

// Determines if the pipeline build was successful
func isPipelineSuccessful() bool {
	status := os.Getenv("BITRISEIO_PIPELINE_BUILD_STATUS")
	log.Debugf("Pipeline Success status: %s\n", status)
	return status == "succeeded" || status == "succeeded_with_abort" || status == ""
}

// Determines if the pipeline build was successful
func isWorkflowSuccessful() bool {
	status := os.Getenv("BITRISE_BUILD_STATUS")
	log.Debugf("Workflow Success status: %s\n", status)
	return status == "0"
}

// selectValue chooses the right value based on the result of the build.
func selectValue(success bool, ifSuccess, ifFailed string) string {
	if success {
		return ifSuccess
	}
	if ifFailed != "" {
		return ifFailed
	}
	return ifSuccess
}

func NewCard(c Config) TeamsMessage {
    success := isPipelineSuccessful() && isWorkflowSuccessful()

    log.Debugf("Success status: %s\n", success)

    card := adaptivecard.NewCard()
    card.Type = "AdaptiveCard"
    card.Schema = "http://adaptivecards.io/schemas/adaptive-card.json"
    card.Version = "1.5"

    // Create style depending on build status
    statusBanner := adaptivecard.NewContainer()
    headline := adaptivecard.NewTextBlock("", false)
    headline.Size = "large"
    headline.Weight = "bolder"
    headline.Style = "heading"

    statusBanner.Style = selectValue(success, c.CardStyle, c.CardStyleOnError)
    headline.Text = selectValue(success, c.CardHeadline, c.CardHeadlineOnError)

    statusBanner.Spacing = "None"
    statusBanner.Separator = true
    statusBanner.Items = append(statusBanner.Items, headline)
    card.Body = append(card.Body, adaptivecard.Element(statusBanner))

    // Main Section
    mainContainer := adaptivecard.NewContainer()
    mainContainer.Style = "default"
    mainContainer.Spacing = "medium"

    title := selectValue(success, c.Title, c.TitleOnError)
    if title != "" {
    	mainContainer.Items = append(mainContainer.Items, adaptivecard.NewTextBlock(title, true))
    }

    if c.AuthorName != "" {
    	mainContainer.Items = append(mainContainer.Items, adaptivecard.NewTextBlock(c.AuthorName, false))
    }

    if c.Subject != "" {
    	mainContainer.Items = append(mainContainer.Items, adaptivecard.NewTextBlock(c.Subject, true))
    }

    // Facts
    factSet := adaptivecard.NewFactSet()
    for _, fact := range parsesFacts(c.Fields) {
    	err := factSet.AddFact(fact)
    	if err != nil {
    		log.Errorf("Could not add fact to factset %v", err)
    	}
    }
    if len(factSet.Facts) > 0 {
    	mainContainer.Items = append(mainContainer.Items, adaptivecard.Element(factSet))
    }

    if len(mainContainer.Items) > 0 {
    	card.Body = append(card.Body, adaptivecard.Element(mainContainer))
    }

    // Images
    imageContainer := parsesImages(selectValue(success, c.Images, c.ImagesOnError))
    if len(imageContainer.Items) > 0 {
    	card.Body = append(card.Body, adaptivecard.Element(imageContainer))
    }

    // Actions (Buttons)
    actions := parsesActions(selectValue(success, c.Buttons, c.ButtonsOnError))
    if len(actions.Actions) > 0 {
    	card.Body = append(card.Body, actions)
    }

    card.MSTeams.Width = "Full"

    return TeamsMessage{
    	Type: "message",
    	Attachments: []Attachment{
    		{
    			ContentType: "application/vnd.microsoft.card.adaptive",
    			Content:     card,
    		},
        },
    }
}

func parsesFacts(s string) (fs []adaptivecard.Fact) {
	for _, p := range pairs(s) {
		fs = append(fs, adaptivecard.Fact{Title: p[0], Value: p[1]})
	}
	return
}

func parsesImages(s string) (container adaptivecard.Container) {
	container = adaptivecard.NewContainer()
	for _, p := range pairs(s) {

		image := adaptivecard.Element{
			URL:  p[1],
			Type: "Image",
			Size: "large",
		}

		err := container.AddElement(false, image)
		if err != nil {
			log.Errorf("Could not add image %v", err)
		}
	}
	return container
}

func parsesActions(s string) (as adaptivecard.Element) {
	as = adaptivecard.NewActionSet()
	for _, p := range pairs(s) {
		action, _ := adaptivecard.NewActionOpenURL(p[1], p[0])
		as.Actions = append(as.Actions, action)
	}

	return as
}

// pairs slices every lines in s into two substrings separated by the first pipe
// character and returns a slice of those pairs.
func pairs(s string) [][2]string {
	var ps [][2]string
	for _, line := range strings.Split(s, "\n") {
		a := strings.SplitN(line, "|", 2)
		if len(a) == 2 && a[0] != "" && a[1] != "" {
			ps = append(ps, [2]string{a[0], a[1]})
		}
	}
	return ps
}

// PostCard sends the given adaptive card to configured webhook
func PostCard(conf Config, msg TeamsMessage) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.Debugf("Post Json Data: %s\n", b)

	url := string(conf.WebhookURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request: %s", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server error: %s, failed to read response: %s", resp.Status, err)
		}
		return fmt.Errorf("server error: %s, response: %s", resp.Status, body)
	}

	return nil
}

func main() {
	var conf Config
	if err := stepconf.Parse(&conf); err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(conf)
	log.SetEnableDebugLog(conf.Debug)

	msg := NewCard(conf)
	if err := PostCard(conf, msg); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Donef("\nMessage successfully sent! 🚀\n")
}
