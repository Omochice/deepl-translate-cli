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
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli/v2"
)

type Setting struct {
	AuthKey    string `json:"-"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type Response struct {
	Translations []Translation
}

type Translation struct {
	DetectedSourceLaguage string `json:"detected_source_language"`
	Text                  string `json:"text"`
}

func LoadSettings(c *cli.Context) (Setting, error) {
	var setting Setting

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return setting, err
	}

	configPath := filepath.Join(homeDir, ".config", "deepl-translation", "setting.json")
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		errStr := fmt.Errorf("Not exists such file. %s\n\tauto make it, please write it. ", configPath)
		err := InitializeConfigFile(configPath)
		if err != nil {
			return setting, err
		}
		return setting, errStr
	}
	if err := json.Unmarshal(bytes, &setting); err != nil {
		return setting, err
	}
	setting.AuthKey = os.Getenv("DEEPL_TOKEN")
	if setting.AuthKey == "" {
		return setting, fmt.Errorf("No deepl token is set.")
	}

	if c.String("source_lang") != "" {
		setting.SourceLang = c.String("source_lang")
	}
	if c.String("target_lang") != "" {
		setting.TargetLang = c.String("target_lang")
	}

	if (setting.SourceLang == setting.TargetLang) && (setting.SourceLang == "FILLIN" || setting.TargetLang == "FILLIN") {
		return setting, fmt.Errorf("Invalid source_lang and target_lang\n\tCheck %s.", configPath)
	}

	return setting, nil
}

func InitializeConfigFile(ConfigPath string) error {
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
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

func Translate(Text string, setting Setting) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", setting.AuthKey)
	params.Add("source_lang", setting.SourceLang)
	params.Add("target_lang", setting.TargetLang)
	params.Add("text", Text)
	baseUrl := "https://api-free.deepl.com/v2/translate"
	resp, err := http.PostForm(baseUrl, params)

	results := []string{}
	if err != nil {
		return results, err
	}

	translateResponse, err := ParseResponse(resp)
	if err != nil {
		return []string{}, err
	}
	for _, translated := range translateResponse.Translations {
		results = append(results, translated.Text)
	}

	return results, err

}

func ParseResponse(resp *http.Response) (Response, error) {
	var responseJson Response
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return responseJson, err
	}
	err = json.Unmarshal(body, &responseJson)
	return responseJson, err
}

func main() {
	app := &cli.App{
		Name:      "deepl",
		Usage:     "Translate sentences.",
		UsageText: "deepl [-s|-t] <inputfile | --stdin> ",

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
		},
		Action: func(c *cli.Context) error {
			setting, err := LoadSettings(c)
			if err != nil {
				return err
			}

			var rawSentense string
			if c.Bool("stdin") {
				if terminal.IsTerminal(syscall.Stdin) {
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

			translateds, err := Translate(rawSentense, setting)
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
