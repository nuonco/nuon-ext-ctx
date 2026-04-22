# nuon-ext-ctx

A [nuon CLI](https://github.com/nuonco/nuon) extension for switching between multiple CLI configurations, inspired by [kubectx](https://github.com/ahmetb/kubectx).

## How it works

`nuon ctx` manages named configurations stored in `~/.config/nuon/contexts/`. The active configuration is a symlink at `~/.nuon` pointing to one of these files. The nuon CLI reads and writes through the symlink transparently.

## Usage

```
USAGE:
  nuon ctx                       : list the contexts
  nuon ctx <NAME>                : switch to context <NAME>
  nuon ctx -                     : switch to the previous context
  nuon ctx -c, --current         : show the current context name
  nuon ctx <NEW_NAME>=<NAME>     : rename context <NAME> to <NEW_NAME>
  nuon ctx <NEW_NAME>=.          : rename current-context to <NEW_NAME>
  nuon ctx -u, --unset           : unset the current context
  nuon ctx -d <NAME> [<NAME...>] : delete context(s) ('.' for current-context)
  nuon ctx -s, --save <NAME>     : save current ~/.nuon as a named context
  nuon ctx -h, --help            : show help
  nuon ctx -V, --version         : show version
```

## Getting started

If you already have a `~/.nuon` config file, save it as a named context:

```bash
nuon ctx -s production
```

This moves the file to `~/.config/nuon/contexts/production` and creates a symlink at `~/.nuon`.

Then log in with different credentials and save another context:

```bash
nuon login
nuon ctx -s staging
```

Now switch between them:

```bash
nuon ctx production
nuon ctx staging
nuon ctx -        # switch back to previous
```

## Building

```bash
make build
```
