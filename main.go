package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unkdown"
	buildBy = "unkdown"
)

func getVersion() string {
	if version != "" {
		return version
	}
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return i.Main.Version
}

type DeepL interface {
	Translate(Text string, setting Setting) ([]string, error)
}

type DeepLClient struct {
	Endpoint string
	AuthKey  string
}

func (c *DeepLClient) Translate(text string, sourceLang string, targetLang string) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)
	params.Add("source_lang", sourceLang)
	params.Add("target_lang", targetLang)
	params.Add("text", text)
	resp, err := http.PostForm(c.Endpoint, params)

	if err := ValidateResponse(resp); err != nil {
		return []string{}, err
	}
	parsed, err := ParseResponse(resp)
	if err != nil {
		return []string{}, err
	}
	r := []string{}
	for _, translated := range parsed.Translations {
		r = append(r, translated.Text)
	}
	return r, nil
}

var KnownErrors = map[int]string{
	400: "Bad request. Please check error message and your parameters.",
	403: "Authorization failed. Please supply a valid auth_key parameter.",
	404: "The requested resource could not be found.",
	413: "The request size exceeds the limit.",
	414: "The request URL is too long. You can avoid this error by using a POST request instead of a GET request, and sending the parameters in the HTTP body.",
	429: "Too many requests. Please wait and resend your request.",
	456: "Quota exceeded. The character limit has been reached.",
	503: "Resource currently unavailable. Try again later.",
	529: "Too many requests. Please wait and resend your request.",
} // this from https://www.deepl.com/docs-api/accessing-the-api/error-handling/

func ValidateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var data map[string]interface{}
		baseErrorText := fmt.Sprintf("Invalid response [%d %s]",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
		if t, ok := KnownErrors[resp.StatusCode]; ok {
			baseErrorText += fmt.Sprintf(" %s", t)
		}
		e := json.NewDecoder(resp.Body).Decode(&data)
		if e != nil {
			return fmt.Errorf("%s", baseErrorText)
		} else {
			return fmt.Errorf("%s, %s", baseErrorText, data["message"])
		}
	}
	return nil
}

type Setting struct {
	AuthKey    string `json:"-"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	IsPro      bool   `json:"-"`
}

type DeepLResponse struct {
	Translations []Translated
}

type Translated struct {
	DetectedSourceLaguage string `json:"detected_source_language"`
	Text                  string `json:"text"`
}

func LoadSettings(setting Setting, automake bool) (Setting, error) {
	if setting.AuthKey == "" {
		return setting, fmt.Errorf("No deepl token is set.")
	}

	if setting.TargetLang == "" || setting.SourceLang == "" {
		homeDir, err := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "deepl-translate-cli", "setting.json")
		// if eigher is not set, load file.
		if err != nil {
			return setting, err
		}

		bytes, err := ioutil.ReadFile(configPath)
		if err != nil {
			errStr := fmt.Errorf("Not exists such file. %s\n\tauto make it, please write it. ", configPath)
			if automake {
				err := InitializeConfigFile(configPath)
				if err != nil {
					return setting, err
				}
			}
			return setting, errStr
		}
		if err := json.Unmarshal(bytes, &setting); err != nil {
			return setting, fmt.Errorf("%s (occurred while loading setting.json)", err.Error())
		}
		if setting.SourceLang == "FILLIN" || setting.TargetLang == "FILLIN" {
			return setting, fmt.Errorf("Did write config file? (%s)", configPath)
		}
	}
	if setting.SourceLang == setting.TargetLang {
		return setting, fmt.Errorf("Equal source lang(%s) and target lang(%s)", setting.SourceLang, setting.TargetLang)
	}
	return setting, nil
}

func InitializeConfigFile(ConfigPath string) error {
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0644); err != nil {
		return err
	}

	initSetting := Setting{
		SourceLang: "FILLIN",
		TargetLang: "FILLIN",
	}

	out, err := os.Create(ConfigPath)
	if err != nil {
		return err
	}
	defer out.Close()

	decoded, err := json.MarshalIndent(initSetting, "", "  ")
	if err != nil {
		return err
	}

	out.Write(([]byte)(decoded))
	return nil
}

func getUrl(IsPro bool) string {
	if IsPro {
		return "https://api.deepl.com/v2/translate"
	}
	return "https://api-free.deepl.com/v2/translate"
}

func ParseResponse(resp *http.Response) (DeepLResponse, error) {
	var responseJson DeepLResponse
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		err := fmt.Errorf("%s (occurred while parse response)", err.Error())
		return responseJson, err
	}
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		err := fmt.Errorf("%s (occurred while parse response)", err.Error())
		return responseJson, err
	}
	return responseJson, err
}

func main() {
	app := &cli.App{
		Name:      "deepl-translate-cli",
		Usage:     "Translate sentences.",
		UsageText: "deepl-translate-cli [-s|-t] <inputfile | --stdin> ",
		Version: fmt.Sprintf(
			"%s (rev %s) [%s %s %s] [build at %s by %s]",
			getVersion(),
			commit,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			date,
			buildBy,
		),
		Authors: []*cli.Author{
			{
				Name: "Omochice",
			},
		},

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "stdin",
				Usage: "use stdin.",
			},
			&cli.StringFlag{
				Name:    "source_lang",
				Aliases: []string{"s"},
				Usage:   "Source `LANG`",
			},

			&cli.StringFlag{
				Name:    "target_lang",
				Aliases: []string{"t"},
				Usage:   "Target `LANG`",
			},
			&cli.BoolFlag{
				Name:  "pro",
				Usage: "use pro plan's endpoint",
			},
		},
		Action: func(c *cli.Context) error {
			setting, err := LoadSettings(Setting{
				SourceLang: c.String("source_lang"),
				TargetLang: c.String("target_lang"),
				AuthKey:    os.Getenv("DEEPL_TOKEN"),
				IsPro:      c.Bool("pro"),
			}, true)
			if err != nil {
				return err
			}

			var rawSentense string
			if c.Bool("stdin") {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					// is not pipe
					fmt.Scan(&rawSentense)
				} else {
					// is pipe
					pipeIn, err := ioutil.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					rawSentense = string(pipeIn)
				}
			} else {
				if c.NArg() == 0 {
					return fmt.Errorf("No filename is set. And `--stdin` option is not set.\nEither must be set.")
				}
				f, err := os.Open(c.Args().First())
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				rawSentense = string(b)
			}

			client := DeepLClient{
				Endpoint: getUrl(c.Bool("pro")),
				AuthKey:  setting.AuthKey,
			}
			translateds, err := client.Translate(rawSentense, setting.SourceLang, setting.TargetLang)
			if err != nil {
				return err
			}
			for _, translated := range translateds {
				fmt.Print(translated)
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
