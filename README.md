# npm-trusted-publishing-tool

Onboards CircleCI projects to [npm trusted publishing](https://docs.npmjs.com/trusted-publishers)
by orchestrating the `circleci` and `npm` CLIs. For users with many packages to configure.

## Requirements

- Go `1.26.4` or later (only needed to build from source)
- `circleci` CLI pre-release (tested with `1.0.38499-pre`), authenticated (`circleci settings set token <token>` or `CIRCLE_TOKEN`). Grab a build from the [releases page](https://github.com/CircleCI-Public/circleci-cli/releases).
- Node.js `22.14.0` or later
- `npm` `11.15.0` or later
- A valid npm token with **write** access and **2FA enabled** (the first `npm trust` call prompts for your 2FA code)

## Download

Download the binary for your platform from the
[releases page](https://github.com/CircleCI-Public/npm-trusted-publishing-tool/releases),
extract the archive, and put `npm-trusted-publishing-tool` on your `PATH`.

## Build

```sh
go build -o npm-trusted-publishing-tool .
```

Optionally, install it onto your `PATH` (drops the binary in `$(go env GOPATH)/bin`,
typically `~/go/bin`):

```sh
go install .
```

## Usage

Run in a single repo (uses the git remote and `./package.json`):

```sh
./npm-trusted-publishing-tool
```

Process many projects from a list file — each line is `package-name,project-slug`:

```
# projects.txt
@acme/widget,gh/acme/widget
@acme/gadget,gh/acme/gadget
```

```sh
./npm-trusted-publishing-tool --projects projects.txt
```

For each project the tool resolves the CircleCI project, lets you pick a pipeline
definition and any contexts (press `/` to filter), asks once which publishing
permissions to grant, then runs `npm trust circleci`.

npm prints its prompts and results directly to your terminal. The **first**
`npm trust` call requires two-factor authentication via the browser: npm shows
an auth URL and waits for you to press ENTER. On that page you can choose to skip
2FA for the next 5 minutes, after which subsequent projects are configured
without prompting (about 80 packages fit in that window).

Add `--dry-run` to pass `--dry-run` to `npm trust` and make no changes (no 2FA).

Projects are grouped by organization and processed org-by-org. Contexts are
organization-scoped, so within an org you're offered to reuse the selected
contexts for the rest of that org; when the org changes you're prompted to
choose contexts again. Publishing permissions are asked once for the whole run.

## License

[MIT](LICENSE).
