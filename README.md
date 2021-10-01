[![go-test](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/ci.yml)
[![goreleaser](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml/badge.svg)](https://github.com/Omochice/deepl-translate-cli/actions/workflows/autorelease.yml)

# ✍️Deepl translate cli

![sampleMovie](https://i.gyazo.com/907bae0779d11c324576ee7777768312.gif)

## Installation

1. Download zipped file from [Releases](https://github.com/Omochice/deepl-translate-cli/releases).

2. Unzip downloaded file.

3. Move executable file into directory in PATH. (like `$HOME/.local/bin/`)


## Installation

### By [github release page](https://github.com/Omochice/deepl-translate-cli/releases)
1. Get deepl access token. See [here](https://www.deepl.com/docs-api).

2. Set access token as `DEEPL_TOKEN`

    ex. in `Bash`.

    ```bash
    export DEEPL_TOKEN <YOUR TOKEN>
    ```

3. Make configure file in `<user home directory>/.config/deepl-translate-cli/setting.json`.

    If run command without existing setting file, auto make it.

    For write setting file, see [this page](https://www.deepl.com/docs-api/translating-text/request/).

### By `go install`
```sh
go install github.com/Omochice/deepl-translate-cli@latest
```

## Usage

- If you want translate from existing file.
    ```console
    $ deepl-translate-cli <text.txt>
    ```

- If you want use stdin.
    - with pipe
        ```console
        $ echo "hello" | deepl-translate-cli --stdin
        ```
    - with input
        ```console
        $ deepl-translate-cli --stdin
        <input text that wanted translate> <Enter>
        ```
