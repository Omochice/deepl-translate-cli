[![go-test](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml)
[![goreleaser](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml/badge.svg)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml)

# ✍️Deepl translate cli

![sampleMovie](https://i.gyazo.com/09a4801d44e85980f83666dceda0166e.gif)


## Installation

### By [github release page](https://github.com/Omochice/deepl-translate-cli/releases)

1. Download zipped file from [Releases](https://github.com/Omochice/deepl-translate-cli/releases).

2. Unzip downloaded file.

3. Move executable file into directory in PATH. (like `$HOME/.local/bin/`)

### By `go install`
```sh
go install github.com/Omochice/deepl-translate-cli@latest
```

## Usage

1. Get deepl access token. See [here](https://www.deepl.com/docs-api).

2. Set access token as `DEEPL_TOKEN`

    ex. in `Bash`.

    ```bash
    export DEEPL_TOKEN <YOUR TOKEN>
    ```

3. On the first run, if `<user home directory>/.config/deepl-translate-cli/setting.json` does not exist, make it automatically. 

    The format of setting file is below.
    ```json
    {
      "source_lang": "FILLIN",
      "target_lang": "FILLIN"
    }
    ```
    For write setting file, see [this page](https://www.deepl.com/docs-api/translating-text/request/).



4. If file path is not specified, load text from STDIN.

    Currentry, only one path can be specified as argument.



- If you want to use `source_lang`/`target_lang` without using setting file, try to use `--source_lang (-s)` or `target_lang (-t)` argument.
    
    ```console
    cat <text.txt> | deepl-translate-cli --source_lang ES --target_lang DE
    ```

- If you use Pro plan, use `--pro` flag to switch endpoint URL.
    
    _this feature is not tested because I use free plan._
    
    ```console
    cat <text.txt> | deepl-translate-cli --pro
    ```
