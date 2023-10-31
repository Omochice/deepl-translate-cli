// Original code by @Omochice <https://github.com/Omochice/deepl-translate-cli>
// With some extra tweaks by Gwyneth Llewelyn <https://gwynethllewelyn.net>
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/Omochice/deepl-translate-cli/deepl"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

// These strings will be retrieved by a call to debug.ReadBuildInfo().
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	buildBy = "unknown"
)

// Shows the version embedded in the binary.
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
	IsPro      bool   `json:"-"`
}

// Open the settings file, or, if it doesn't exist, create it first.
func LoadSettings(setting Setting, automake bool) (Setting, error) {
	if setting.AuthKey == "" {
		return setting, fmt.Errorf("no DeepL token is set")
	}

	if setting.TargetLang == "" || setting.SourceLang == "" {
		homeDir, err := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "deepl-translate-cli", "setting.json")
		// if either is not set, load file.
		if err != nil {
			return setting, err
		}

		bytes, err := os.ReadFile(configPath)
		if err != nil {
			errStr := fmt.Errorf("file does not exist. %s\n\tauto make it, please write it", configPath)
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
			return setting, fmt.Errorf("did write config file? (%s)", configPath)
		}
	}
	if setting.SourceLang == setting.TargetLang {
		return setting, fmt.Errorf("cannot have identical source lang(%s) and target lang(%s)", setting.SourceLang, setting.TargetLang)
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

func main() {
	app := &cli.App{
		Name:      "deepl-translate-cli",
		Usage:     "Translate sentences.",
		UsageText: "deepl-translate-cli [-s|-t] <inputfile>",
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
			&cli.StringFlag{
				Name:    "source_lang",
				Aliases: []string{"s"},
				Usage:   "Set source language without using the settings file",
			},

			&cli.StringFlag{
				Name:    "target_lang",
				Aliases: []string{"t"},
				Usage:   "Set target language without using the settings file",
			},
			&cli.BoolFlag{
				Name:  "pro",
				Usage: "Use pro plan's endpoint",
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

			var rawSentence string
			if c.NArg() == 0 {
				// no filename path passed; read from STDIN (TTY or pipe)
				if isatty.IsTerminal(os.Stdin.Fd()) {
					// is not pipe (i.e. TTY)
					fmt.Scan(&rawSentence)
				} else {
					// is pipe
					pipeIn, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					rawSentence = string(pipeIn)
				}
			} else {
				if c.NArg() >= 2 {
					return fmt.Errorf("cannot specify multiple file paths")
				}
				f, err := os.Open(c.Args().First())
				if err != nil {
					return err
				}
				b, err := io.ReadAll(f)
				if err != nil {
					return err
				}
				rawSentence = string(b)
			}

			client := deepl.DeepLClient{
				Endpoint: deepl.GetEndpoint(c.Bool("pro")),
				AuthKey:  setting.AuthKey,
			}
			translateds, err := client.Translate(rawSentence, setting.SourceLang, setting.TargetLang)
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
