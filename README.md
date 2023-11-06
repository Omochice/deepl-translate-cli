[![go-test](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml)
[![goreleaser](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml/badge.svg)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml)

# ✍️ DeepL Translate CLI

![sampleMovie](https://i.gyazo.com/09a4801d44e85980f83666dceda0166e.gif)

## Installation

### Via the [GitHub release page](https://github.com/Omochice/deepl-translate-cli/releases)

1. Download zipped file from [Releases](https://github.com/Omochice/deepl-translate-cli/releases).

2. Unzip downloaded file.

3. Move executable file into a directory in `PATH`. (like `$HOME/.local/bin/`)

### By `go install`

```console
go install github.com/Omochice/deepl-translate-cli@latest
```

## Usage

1. First, [get a DeepL access token](https://www.deepl.com/docs-api). It looks like a [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) with the characters `:fx` appended to it.

2. Assign the access token to the `DEEPL_TOKEN` environment variable.

    e.g., in `bash`:

    ```console
    export DEEPL_TOKEN=<YOUR DEEPL API TOKEN>
    ```

3. On the first run, if `<USER HOME DIRECTORY>/.config/deepl-translate-cli/setting.json` does not exist, it gets automatically created.

    The format of the settings file is as shown below:

    ```json
    {
    	"source_lang": "FILLIN",
    	"target_lang": "FILLIN"
    }
    ```

    For all existing languages that can be translated, as well as their identifying tags, see [this page](https://www.deepl.com/docs-api/translating-text/request/). You can also query the server directly:
	
	```console
	deepl-translate-cli languages
	
	```

4. If the filename path is not specified, text is read from `STDIN`.

    Currently, only one path can be specified as argument.

-   If you want to select `source_lang`/`target_lang` _without_ using the settings file, you can use the command-line parameters `--source_lang (-s)` and `target_lang (-t)` instead.

    ```console
    cat <text.txt> | deepl-translate-cli --source_lang ES --target_lang DE
    ```

-   If you are a Pro plan user, switch to the correct endpoint URL with the `--pro` flag.

    _**Note**: This feature has not been tested, because the developers only have a free plan._

    ```console
    cat <text.txt> | deepl-translate-cli --pro

    ```

## Shell autocompletion (⚠️ experimental)

Under the `autocomplete` folder are three scripts to enable auto-completion (for `bash`, `zsh`, and PowerShell). To use these, do the following (the example is for `bash`):

```console
PROG=deepl-translate-cli source autocomplete/bash_autocomplete
```

## ⚠️ Warning! ⚠️

If you run the tests, these may actually use your API Token, and consume some of your monthly credits!

## Building

If you wish to embed the build's author in the executable binary (to distinguish _your_ build from someone else's), you can build this with

```console
go build -ldflags "-X main.TheBuilder=<YOUR NAME HERE>"
```

## Disclaimer

None of the developers are affiliated with [DeepL](https://www.deepl.com/) and this code should not be considered to represent an endorsement by DeepL or any of its affiliates, partners or subsidiaries. It is released in the hope that it might be helpful to the Go programming community (which lacks official support by DeepL at the time of writing), without any warranty whatsoever (see [LICENSE] for more information).
