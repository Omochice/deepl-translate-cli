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
	"time"

	"github.com/Omochice/deepl-translate-cli/deepl"
	"github.com/lmorg/readline"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

// versionInfoType holds the relevant information for this build.
// It is meant to be used as a cache.
type versionInfoType struct {
	version		string		// Runtime version.
	commit  	string		// Commit revision number.
	dateString 	string		// Commit revision time (as a RFC3339 string).
	date		time.Time	// Same as before, converted to a time.Time, because that's what the cli package uses.
	builtBy 	string		// User who built this (see note).
	goOS		string		// Operating system for this build (from runtime).
	goARCH		string		// Architecture, i.e., CPU type (from runtime).
	goVersion	string		// Go version used to compile this build (from runtime).
	init		bool		// Have we already initialised the cache object?
}

// NOTE: I don't know where the "builtBy" information comes from, so, right now, it gets injected
// during build time, e.g. `go build -ldflags "-X main.TheBuilder=gwyneth"` (gwyneth 20231103)

var (
	versionInfo versionInfoType	// cached values for this build.
	TheBuilder string			// to be overwritten via the linker command `go build -ldflags "-X main.TheBuilder=gwyneth"`.
	debugLevel int				// verbosity/debug level.
)

// Initialises the versionInfo variable.
func initVersionInfo() error {
	if versionInfo.init {
		// already initialised, no need to do anything else!
		return nil
	}
	// get the following entries from the runtime:
	versionInfo.goOS		= runtime.GOOS
	versionInfo.goARCH		= runtime.GOARCH
	versionInfo.goVersion	= runtime.Version()

	// attempt to get some build info as well:
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("no valid build information found")
	}
	versionInfo.version = buildInfo.Main.Version

	// Now dig through settings and extract what we can...

	var vcs, rev string // Name of the version control system name (very likely Git) and the revision.
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
			case "vcs":
				vcs = setting.Value
			case "vcs.revision":
				rev = setting.Value
			case "vcs.time":
				versionInfo.dateString = setting.Value
		}
	}
	versionInfo.commit = "unknown"
	if vcs != "" {
		versionInfo.commit = vcs
	}
	if rev != "" {
		versionInfo.commit += " [" + rev + "]"
	}
	// attempt to parse the date, which comes as a string in RFC3339 format, into a date.Time:
	var parseErr error
	if versionInfo.date, parseErr = time.Parse(versionInfo.dateString, time.RFC3339); parseErr != nil {
		// Note: we can safely ignore the parsing error: either the conversion works, or it doesn't, and we
		// cannot do anything about it... (gwyneth 20231103)
		// However, the AI revision bots dislike this, so we'll assign the current date instead.
		versionInfo.date = time.Now()

		if debugLevel > 1 {
			fmt.Fprintf(os.Stderr, "date parse error: %v", parseErr)
		}
	}

	// NOTE: I have no idea where the "builtBy" info is supposed to come from;
	// the way I do it is to force the variable with a compile-time option. (gwyneth 20231103)
	versionInfo.builtBy = TheBuilder

	return nil
}

// Internal settings, to be filled by LoadSettings(), and which gets saved to a file to
// be reused on subsequent calls.
// NOTE: This might become utterly different if we implement settings stored via
// the github.com/urfave/cli-altsrc package. (gwyneth 20231103)
type Setting struct {
	AuthKey    			string	`json:"-"`						// API token, looks like a UUID with ":fx".
	SourceLang 			string	`json:"source_lang"`
	TargetLang 			string	`json:"target_lang"`
	LanguagesType		string	`json:"type"`					// For the "languages" utility call, either "source" or "target".
	IsPro      			bool	`json:"-"`
	TagHandling			string	`json:"tag_handling"`			// "xml", "html".
	SplitSentences		string	`json:"split_sentences"`		// "0", "1", "norewrite".
	PreserveFormatting	string	`json:"preserve_formatting"`	// "0", "1".
	OutlineDetection	int		`json:"outline_detection"`		// Integer; 0 is default.
	NonSplittingTags	string	`json:"non_splitting_tags"`		// List of comma-separated XML tags.
	SplittingTags		string	`json:"splitting_tags"`			// List of comma-separated XML tags.
	IgnoreTags			string	`json:"ignore_tags"`			// List of comma-separated XML tags.
	Debug				int		`json:"debug"`					// Debug/verbosity level, 0 is no debugging.
}

// Open the settings file, or, if it doesn't exist, create it first.
// TODO: Probably change all this to use github.com/urfave/cli-altsrc instead.
func LoadSettings(setting Setting, automake bool) (Setting, error) {
	if setting.AuthKey == "" {
		return setting, fmt.Errorf("no DeepL token is set; use the environment variable `DEEPL_TOKEN` to set it")
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
			errStr := fmt.Errorf("settings file does not exist. %s\n\tIt was autogenerated, please edit it to reflect your preferences", configPath)
			if automake {
				err := InitializeConfigFile(configPath)
				if err != nil {
					return setting, err
				}
			}
			return setting, errStr
		}
		if err := json.Unmarshal(bytes, &setting); err != nil {
			return setting, fmt.Errorf("%s (occurred while loading `setting.json`)", err.Error())
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

// Attempts to create the directory for the configuration file, with a minimalist configurztion if successful.
// If creating the directory (or the file within) fails, then abort and return error.
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

// TODO: Try to use "github.com/urfave/cli/v3" in the future...
// TODO: @urfave has his own library to deal with configuration files, cli-altsrc.
//       It's obscure and sparsely documented (see ).
//       But it's probably far more flexible than the simplistic scheme used here. (gwyneth 20231103)
func main() {
	// Global settings for this cli app.
	var setting Setting
	// DeepL Token, usually coming from the environmant variable `DEEPL_TOKEN`.
	var deeplToken string

	// Set up the version/runtime/debug-related variables, and cache them:
	initVersionInfo()

	// Test if the authentication can work or not, depending if we got the token
	// set as an environment variable.
	deeplToken, ok := os.LookupEnv("DEEPL_TOKEN")
	if !ok {
		fmt.Fprintln(os.Stderr, "Please set first your DeepL authentication key using the environment variable DEEPL_TOKEN.")
		os.Exit(1)	// NOTE: the cli.Exit() function cannot be used here, because cli is not initialized yet.
		// return cli.Exit(fmt.Sprintln("Please set first your DeepL authentication key using the environment variable DEEPL_TOKEN."), 1)
	}
	// Generic error variable to work around scoping issues.
	var err error
	// Configure all settings from the very start, because we need the authkey & endpoint
	// for all other calls, not just translations.
	setting, err = LoadSettings(
		Setting{
			AuthKey: deeplToken,
		},
		true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot init settings, error was: %q", err)
		os.Exit(1)
		// return cli.Exit(fmt.Sprintf("cannot init settings, error was: %q", err), 1)
	}

	// start app
	app := &cli.App{
		Name:      "deepl-translate-cli",
		Usage:     "Translate sentences, using the DeepL API.",
		UsageText: "deepl-translate-cli [-s|-t][--pro] trans [--tag_handling [xml|html]] <inputfile>\ndeepl-translate-cli usage\ndeepl-translate-cli languages [--type=[source|target]]\ndeepl-translate-cli glossary-language-pairs",
		Version: fmt.Sprintf(
			"%s (rev %s) [%s %s %s] [build at %s by %s]",
			versionInfo.version,
			versionInfo.commit,
			versionInfo.goOS,
			versionInfo.goARCH,
			versionInfo.goVersion,
			versionInfo.dateString,		// Date as string in RFC3339 notation.
			versionInfo.builtBy,		// see note at the top...
		),
		DefaultCommand: "translate",	// to avoid brealing compatibility with earlier versions.
		EnableBashCompletion: true,
		Compiled: versionInfo.date,		// Converted from RFC333
		Authors: []*cli.Author{
			{
				Name: "Omochice",
				Email: "somewhere@here.jp",
			},
			{
				Name: "Gwyneth Llewelyn",
				Email: "gwyneth.llewelyn@gwynethllewelyn.net",
			},
		},
		Copyright: "© 2021-2023 by Omochice. All rights reserved. Freely distributed under a MIT license.\nThis software is not affiliated nor endorsed by DeepL SE.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "source_lang",
				Aliases: []string{"s"},
				Usage:   "Set source language without using the settings file",
				Value:	 "EN",
				Destination:	&setting.SourceLang,
			},
			&cli.StringFlag{
				Name:    "target_lang",
				Aliases: []string{"t"},
				Usage:   "Set target language without using the settings file",
				Value:	 "JA",
				Destination:	&setting.TargetLang,
			},
			&cli.BoolFlag{
				Name:    "pro",
				Usage:   "Use Pro plan's endpoint?",
				Value:   false,
				Destination: &setting.IsPro,
			},
			&cli.BoolFlag{
				Name:	"debug",
				Aliases: []string{"d"},
				Usage:	"Debugging; repeating the flag increases verbosity.",
				Count:	&debugLevel,
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "translate",
				Aliases:     []string{"trans"},
				Usage:       "Basic translation of a set of Unicode strings into another language",
				Description: "Text to be translated.\nOnly UTF-8-encoded plain text is supported. May contain multiple sentences, but the total request body size must not exceed 128 KiB (128 · 1024 bytes).\nPlease split up your text into multiple	calls if it exceeds this limit.",
				Category:	 "Translations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "tag_handling",
						Usage:       "Set to XML or HTML in order to do more advanced parsing (empty means just using the plain text variant)",
						Aliases:     []string{"tag"},
						Value:       "",
						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, v string) error {
							switch v {
								case "xml", "html":
									return nil
								default:
									return fmt.Errorf("tag_handling must be either `xml` or `html` (got: %s)", v)
							}
						},
					},
					&cli.StringFlag{
						Name:        "split_sentences",
						Usage:       "Sets whether the translation engine should first split the input into sentences. For text translations where `tag_handling` is not set to `html`, the default value is `1`, meaning the engine splits on punctuation and on newlines.\nFor text translations where `tag_handling=html`, the default value is `nonewlines`, meaning the engine splits on punctuation only, ignoring newlines.\n\nThe use of `nonewlines` as the default value for text translations where `tag_handling=html` is new behavior that was implemented in November 2022, when HTML handling was moved out of beta.\n\nPossible values are:\n\n * `0` - no splitting at all, whole input is treated as one sentence\n * `1` (default when `tag_handling` is not set to `html`) - splits on punctuation and on newlines\n * `nonewlines` (default when `tag_handling=html`) - splits on punctuation only, ignoring newlines\n\nFor applications that send one sentence per text parameter, we recommend setting `split_sentences` to `0`, in order to prevent the engine from splitting the sentence unintentionally.\n\nPlease note that newlines will split sentences when `split_sentences=1`. We recommend cleaning files so they don't contain breaking sentences or setting the parameter `split_sentences` to `nonewlines`.",
						Aliases:     []string{"split"},
						Value:       "",	// NOTE: default value should depend on `tag_handling`.
						Destination:	&setting.SplitSentences,
						Action: func(c *cli.Context, v string) error {
							switch v {
								case "0", "1", "nonewlines":
									return nil
								default:
									return fmt.Errorf("split_sentences can only be 0, 1, or `nonewlines` (got: %s)", v)
							}
						},
					},
					&cli.StringFlag{
						Name:        "preserve_formatting",
						Usage:       "Sets whether the translation engine should respect the original formatting, even if it would usually correct some aspects. Possible values are:\n * `0` (default)\n * `1`\n\nThe formatting aspects affected by this setting include:\n * Punctuation at the beginning and end of the sentence\n * Upper/lower case at the beginning of the sentence",
						Aliases:     []string{"preserve"},
						Value:       "0",
						Destination: &setting.PreserveFormatting,
						Action: func(c *cli.Context, v string) error {
							switch v {
								case "0", "1":
									return nil
								default:
									return fmt.Errorf("preserve_formatting can only be 0 or 1 (got: %s)", v)
							}
						},
					},
					&cli.IntFlag{
						Name:        "outline_detection",
						Usage:       "The automatic detection of the XML structure won't yield best results in all XML files. You can disable this automatic mechanism altogether by setting the `outline_detection` parameter to `false` and selecting the tags that should be considered structure tags. This will split sentences using the `splitting_tags` parameter.",
						Aliases:     []string{"outline"},
						Value:       0,
						Destination: &setting.OutlineDetection,
					},
					&cli.StringFlag{
						Name:        "non_splitting_tags",
						Usage:       "Comma-separated list of XML tags which never split sentences.",
						Aliases:     []string{"never"},
						//Value:       [""],
						Destination: &setting.NonSplittingTags,
					},
					&cli.StringFlag{
						Name:        "splitting_tags",
						Usage:       "Comma-separated list of XML tags which always cause splits.",
						Aliases:     []string{"always"},
						//Value:       [""],
						Destination: &setting.SplittingTags,
					},
					&cli.StringFlag{
						Name:        "ignore_tags",
						Usage:       "Comma-separated list of XML tags which will always be ignored.",
						Aliases:     []string{"ignore"},
						//Value:       [""],
						Destination: &setting.IgnoreTags,
					},
				},
				Action: func(c *cli.Context) error {
/*
					if c.String("source_lang") != "" {
						setting.SourceLang = c.String("source_lang")
					}
					if c.String("target_lang") != "" {
						setting.TargetLang = c.String("target_lang")
					}
					if c.Bool("pro") {
						setting.IsPro = true
					}
*/
					// TODO(gwyneth): Create constants for debugging levels.
					if debugLevel > 1 {
						fmt.Fprintf(os.Stderr, "Number of args (Narg): %d, c.Args.Len(): %d\n", c.NArg(), c.Args().Len())
					}
					var rawSentence string
					if c.NArg() == 0 {
						// no filename path passed; read from STDIN (TTY or pipe)
						if isatty.IsTerminal(os.Stdin.Fd()) {
							// is not pipe (i.e. TTY)
							// NOTE: This seems not to work very well...(gwyneth 20231101)
							// fmt.Scan(&rawSentence)
							// Replaced it by using a readline (from a library), but it might really be overkill, since
							rl := readline.NewInstance()
							rawSentence, err = rl.Readline()
							if err != nil {
								return err
							}
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
						Endpoint: 			deepl.GetEndpoint(c.Bool("pro")) + "/translate",
						AuthKey:			deeplToken,
						SourceLang:			c.String("source_lang"),
						TargetLang:			c.String("target_lang"),
						LanguagesType:		c.String("type"),
						IsPro:				c.Bool("pro"),
						TagHandling:		c.String("tag_handling"),
						SplitSentences:		c.String("split_sentences"),
						PreserveFormatting:	c.String("preserve_formatting"),
						OutlineDetection:	c.Int("outline_detection"),
						NonSplittingTags:	c.String("non_splitting_tags"),
						SplittingTags:		c.String("splitting_tags"),
						IgnoreTags:			c.String("ignore_tags"),
						Debug:				debugLevel,
					}

					// Simplified call to Translate, now everything is passed via the DeepLClient
					// initialisation.
					translateds, err := client.Translate(rawSentence)
					if err != nil {
						return err
					}
					for _, translated := range translateds {
						fmt.Print(translated)
					}
					return nil
				},
			},
			{
				Name:        "usage",
				Aliases:     []string{"u"},
				Usage:       "Check usage and limits",
				Description: "Retrieve usage information within the current billing period together with the corresponding account limits.",
				Category:	 "Utilities",
				Action: func(c *cli.Context) error {
					client := deepl.DeepLClient{
						Endpoint: deepl.GetEndpoint(c.Bool("pro")) + "/usage",
						AuthKey:  setting.AuthKey,
					}
					s, err := client.Usage()
					if err != nil {
						return err
					}
					fmt.Println(s)
					return nil
				},
			},
			{
				// TODO: make a call to languages and store the valid pairs retrieved,
				// so that we can later validate them. (gwyneth 20231105)
				Name:        "languages",
				// Aliases:     []string{"l"},
				Usage:       "Retrieve supported languages",
				Description: "Retrieve the list of languages that are currently supported for translation, either as source or target language, respectively.",
				Category:	 "Utilities",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "type",
						Usage:       "`TYPE` sets whether source or target languages should be listed. Possible options are:\n`source`: For languages that can be used in the `source_lang` parameter of translate requests.\n`target`: For languages that can be used in the `target_lang` parameter of translate requests.\n",
						Value:       "source",
						DefaultText: "source",
						Action: func(c *cli.Context, v string) error {
							switch v {
								case "source", "target":
									return nil
								default:
									return fmt.Errorf("type must be either `source` or `target` (got: %s)", v)
							}
						},
					},
				},
				Action: func(c *cli.Context) error {
					client := deepl.DeepLClient{
						Endpoint:		deepl.GetEndpoint(c.Bool("pro")) + "/languages",
						AuthKey:  		setting.AuthKey,
						LanguagesType:	c.String("type"),
					}
					s, err := client.Languages()
					if err != nil {
						return err
					}
					fmt.Println(s)
					return nil
				},
			},
			{
				Name:        "glossary-language-pairs",
				// Aliases:     []string{"l"},
				Usage:       "List language pairs supported by glossaries",
				Description: "Retrieve the list of language pairs supported by the glossary feature.",
				Category:	 "Glossary",
				Action: func(c *cli.Context) error {
					client := deepl.DeepLClient{
						Endpoint:	deepl.GetEndpoint(c.Bool("pro")) + "/glossary-language-pairs",
						AuthKey:	setting.AuthKey,
					}
					s, err := client.GlossaryLanguagePairs()
					if err != nil {
						return err
					}
					fmt.Println(s)
					return nil
				},
			},
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
