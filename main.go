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

func LoadSettings(setting Setting) (Setting, error) {
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
			err := InitializeConfigFile(configPath)
			if err != nil {
				return setting, err
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

func Translate(Text string, setting Setting) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", setting.AuthKey)
	params.Add("source_lang", setting.SourceLang)
	params.Add("target_lang", setting.TargetLang)
	params.Add("text", Text)
	endpoint := "https://api-free.deepl.com/v2/translate"
	resp, err := http.PostForm(endpoint, params)

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
		Version: fmt.Sprintf("%s (rev %s) [%s %s %s] [build at %s by %s]",
			getVersion(),
			commit,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			date,
			buildBy),
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
		},
		Action: func(c *cli.Context) error {
			setting, err := LoadSettings(Setting{
				SourceLang: c.String("source_lang"),
				TargetLang: c.String("target_lang"),
				AuthKey:    os.Getenv("DEEPL_TOKEN"),
			})
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
