# ✍️Deepl translate cli

## Installation

1. Download zipped file from [Releases](https://github.com/Omochice/deepl-translate-cli/releases).

2. Unzip downloaded file.

3. Move executable file into directory in PATH. (like `$HOME/.local/bin/`)


## Before using

1. Get deepl access token. See [here](https://www.deepl.com/docs-api).

2. Set access token as `DEEPL_TOKEN`
 ex. in `Bash`.
 ```bash
 export DEEPL_TOKEN <YOUR TOKEN>
 ```

3. Make configure file in `<user home directory>/.config/deepl-translation/setting.json`.

    If run command without existing setting file, auto make it.

    For write setting file, see [this page](https://www.deepl.com/docs-api/translating-text/request/).

## Usage

- If you want translate from existing file.
```console
$ deepl-translation <text.txt>
```

- If you want use stdin.
    - with pipe 
        ```console
        $ echo "hello" | deepl-translation --stdin
        ```
    - with input
        ```console
        $ deepl-translation --stdin
        <input text that wanted translate> <Enter>
        ```
