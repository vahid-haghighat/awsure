package internal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"log"
	"strings"
	"time"
)

var configFileVersion = "1.0.0"

type configuration struct {
	AzureTenantId        string `yaml:"azure_tenant_id"`
	AzureAppIdUri        string `yaml:"azure_app_id_uri"`
	AzureUsername        string `yaml:"azure_username"`
	OktaUsername         string `yaml:"okta_username"`
	RememberMe           bool   `yaml:"remember_me"`
	DefaultRoleArn       string `yaml:"default_role_arn"`
	DefaultDurationHours int    `yaml:"default_duration_hours"`
	Region               string `yaml:"region"`
}

type configurationFile struct {
	Version string                    `yaml:"version"`
	Configs map[string]*configuration `yaml:"configs"`
}

type state struct {
	name     string
	selector string
	handler  func(pg *rod.Page, el *rod.Element, conf *configuration) error
}

type samlResponse struct {
	XMLName   xml.Name
	Assertion samlAssertion `xml:"Assertion"`
}

type samlAssertion struct {
	XMLName            xml.Name
	AttributeStatement samlAttributeStatement
}

type samlAttributeValue struct {
	XMLName xml.Name
	Type    string `xml:"xsi:type,attr"`
	Value   string `xml:",innerxml"`
}

type samlAttribute struct {
	XMLName         xml.Name
	Name            string               `xml:",attr"`
	AttributeValues []samlAttributeValue `xml:"AttributeValue"`
}

type samlAttributeStatement struct {
	XMLName    xml.Name
	Attributes []samlAttribute `xml:"Attribute"`
}

type role struct {
	roleArn      string
	principalArn string
}

var states = []state{
	{
		name:     "username input",
		selector: `input[name="loginfmt"]:not(.moveOffScreen)`,
		handler: func(pg *rod.Page, el *rod.Element, conf *configuration) error {
			var err error
			username := conf.AzureUsername
			if username == "" {
				prompter := Prompter{}
				username, err = prompter.Prompt("Azure Username", username)
				if err != nil {
					return err
				}
			}

			el.MustWaitVisible()
			el.MustSelectAllText().MustInput("")
			el.MustInput(strings.TrimSpace(username))

			sb := pg.MustElement(`input[type=submit]`)

			sb.MustWaitVisible()
			wait := pg.MustWaitRequestIdle()
			sb.MustClick()
			wait()

			pContext := pg.GetContext()
			defer func() {
				pg.Context(pContext)
			}()

			ctx, cancel := context.WithCancel(pContext)
			defer cancel()

			ch := make(chan bool, 1)

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						_, err := pg.Sleeper(rod.NotFoundSleeper).Element("input[name=loginfmt]")
						if err != nil {
							ch <- true
							return
						}
					}
				}
			}()

			go func() {
				_, err := pg.Timeout(20 * time.Second).Race().
					Element("input[name=loginfmt].has-error").
					Element("input[name=loginfmt].moveOffScreen").
					Element("input[name=loginfmt]").Handle(func(e *rod.Element) error {
					return e.WaitInvisible()
				}).Do()
				if err != nil {
					return
				}

				select {
				case <-ctx.Done():
					return
				default:
					ch <- true
					return
				}
			}()

			select {
			case <-ch:
			case <-time.After(25 * time.Second):
			}
			return nil
		},
	},
	{
		name:     "password input",
		selector: `input[name="Password"]:not(.moveOffScreen),input[name="passwd"]:not(.moveOffScreen)`,
		handler: func(pg *rod.Page, el *rod.Element, conf *configuration) error {
			alert, err := pg.Sleeper(rod.NotFoundSleeper).Element(".alert-error")

			if alert != nil && err == nil {
				log.Println(alert.Text())
			}

			prompter := Prompter{}
			password, err := prompter.SensitivePrompt("Azure Password")

			el.MustWaitVisible()
			el.MustSelectAllText().MustInput("")
			el.MustInput(password)

			wait := pg.MustWaitRequestIdle()
			pg.MustElement("span[class=submit],input[type=submit]").MustClick()
			wait()

			time.Sleep(time.Millisecond * 500)
			return nil
		},
	},
	{
		name:     "OKTA username/password input",
		selector: `form:not(.o-form-saving) > div span.okta-form-input-field input[name="identifier"]:not([disabled])`,
		handler: func(pg *rod.Page, el *rod.Element, conf *configuration) error {
			errorSelector := `div.o-form-error-container`
			errorContainer, err := pg.Sleeper(rod.NotFoundSleeper).Element(errorSelector)

			if errorContainer != nil && err == nil {
				t, _ := errorContainer.Text()
				if t != "" {
					fmt.Println(t)
				}
			}

			infoSelector := `div.o-form-info-container`
			infoContainer, err := pg.Sleeper(rod.NotFoundSleeper).Element(infoSelector)
			if infoContainer != nil && err == nil {
				t, _ := infoContainer.Text()
				if t != "" {
					fmt.Println(t)
				}
			}

			prompter := Prompter{}
			username := conf.OktaUsername
			if username == "" {
				username, err = prompter.Prompt("Okta Username", username)
				if err != nil {
					return err
				}
			}

			el.MustWaitVisible()
			el.MustSelectAllText().MustInput("")
			el.MustInput(username)

			password, err := prompter.SensitivePrompt("Okta Password")
			if err != nil {
				return err
			}

			time.Sleep(time.Millisecond * 500)

			pwdEl := pg.MustElement(`input[type="password"]`)
			pwdEl.MustWaitVisible()
			pwdEl.MustSelectAllText().MustInput("")
			pwdEl.MustInput(password)

			time.Sleep(time.Millisecond * 500)

			submitSelector := `input:not([disabled]):not(.link-button-disabled):not(.btn-disabled)[type=submit]`

			btn, err := pg.Sleeper(rod.NotFoundSleeper).Element(submitSelector)
			if err == nil {
				wait := pg.MustWaitRequestIdle()
				btn.MustClick()
				wait()

				pContext := pg.GetContext()
				defer func() {
					pg.Context(pContext)
				}()

				ctx, cancel := context.WithCancel(pContext)
				defer cancel()

				ch := make(chan bool, 1)

				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						default:
							_, err := pg.Sleeper(rod.NotFoundSleeper).Element(submitSelector)
							if err != nil {
								ch <- true
								return
							}
						}
					}
				}()

				go func() {
					_, err := pg.Timeout(20 * time.Second).Race().
						Element(errorSelector + `.o-form-has-errors`).Handle(func(e *rod.Element) error {
						if e != nil {
							t, _ := e.Text()
							if t != "" {
								return errors.New("error returned")
							}
						}
						return nil
					}).
						Element(submitSelector).Handle(func(e *rod.Element) error {
						return e.WaitInvisible()
					}).Do()
					if err != nil {
						return
					}

					select {
					case <-ctx.Done():
						return
					default:
						ch <- true
						return
					}
				}()

				select {
				case <-ch:
				case <-time.After(25 * time.Second):
				}
			}
			return nil
		},
	},
	{
		name:     "OKTA SELECT PUSH Form",
		selector: `div[data-se="okta_verify-push"] > a:not([disabled]):not(.link-button-disabled):not(.btn-disabled)`,
		handler: func(pg *rod.Page, el *rod.Element, conf *configuration) error {
			alert, err := pg.Sleeper(rod.NotFoundSleeper).Element(".infobox-error")

			if alert != nil && err == nil {
				t, _ := alert.Text()
				if t != "" {
					fmt.Println(t)
				}
			}

			btn, err := pg.Sleeper(rod.NotFoundSleeper).Element(`div[data-se="okta_verify-push"] > a:not([disabled]):not(.btn-disabled):not(.link-button-disabled)`)
			if err == nil && btn != nil {
				btn.MustWaitVisible()
				wait := pg.MustWaitRequestIdle()
				btn.MustClick()
				wait()
				time.Sleep(time.Millisecond * 500)
			}
			return nil
		},
	},
	{
		name:     "OKTA DO PUSH Form",
		selector: `a.send-push:not([disabled]):not(.link-button-disabled):not(.btn-disabled)`,
		handler: func(pg *rod.Page, el *rod.Element, conf *configuration) error {
			alert, err := pg.Sleeper(rod.NotFoundSleeper).Element(".infobox-error")

			if alert != nil && err == nil {
				t, _ := alert.Text()
				if t != "" {
					fmt.Println(t)
				}
			}

			btn, err := pg.Sleeper(rod.NotFoundSleeper).Element(`a.send-push:not([disabled]):not(.btn-disabled):not(.link-button-disabled)`)
			if err == nil && btn != nil {
				btn.MustWaitVisible()
				wait := pg.MustWaitRequestIdle()
				btn.MustClick()
				wait()
				time.Sleep(time.Millisecond * 500)
			}
			return nil
		},
	},
}
