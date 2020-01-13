# jty: Jsonnet To Yaml

jty (prounounced "jutty", rhymes with "putty") is a simple utility that does one thing:
it processes many Jsonnet files and emits YAML to specified output files.

While some tools are able to directly consume Jsonnet, generating and committing the resulting YAML
allows authors to modify the Jsonnet with confidence in the resulting output.
This also helps with confidence in reviewing Jsonnet changes.

## What jty does

jty simply accepts input as pairs of /path/to/input.jsonnet and /path/to/output.yml, and in a single process evaluates all the input Jsonnet to generate the corresponding output YAML.
This way, .libsonnet files that are imported more than once are read and evaluated only once.

jty produces human-reader-friendly YAML, unlike `jsonnet -y` which effectively emits JSON, which is also valid YAML.
That is, jty produces:

```yaml
---
numbers:
  - 1
  - 2
  - 3
object:
    that:
        is:
            deeply: nested
```

and `jsonnet -y` produces:

```yaml
---
{
   "numbers": [
      1,
      2,
      3
   ],
   "object": {
      "that": {
         "is": {
            "deeply": "nested"
         }
      }
   }
}
...
```

It also supports `JSONNET_PATH` and the `--jpath`/`-J` flags like the official `jsonnet` command.

If jty still isn't fast enough for your needs,
perhaps [Databricks' SJsonnet](https://databricks.com/blog/2018/10/12/writing-a-faster-jsonnet-compiler.html)
would be a better fit for you.
But SJsonnet also requires a JVM, whereas jty is written and Go and is distributable as a ~10MB standalone binary.

## What jty doesn't do

jty currently does not support setting top-level arguments or external variables.
Support on a global level would be straightforward but not necessarily useful.
But there isn't an obviously intuitive way to supply top-level arguments or external variables on a per-file basis,
so for now, if your Jsonnet requires them, you can use multiple invocations of standard `jsonnet` or you can modify jty.

## Example uses

### Explicit positional arguments

Evaluate in.jsonnet and save the resulting YAML as out.yaml:

    jty in.jsonnet out.yaml

Evaluate multiple .jsonnet files and save the resulting YAML in specific locations:

    jty in1.jsonnet out/1.yaml conf.jsonnet conf.yaml

### Reading from stdin

You can supply a sequence of input file, output file, input file, output file...
to stdin when you use `jty -i`.

Evaluate each .jsonnet file under the current directory,
and save the .yml file adjacent to the .jsonnet file
(e.g. `./app.jsonnet` -> `./app.yml`):

    find . -name '*.jsonnet' \
      -exec bash -c 'for p in "$@"; do
        printf "%s\n%s.yml\n" "$p" "${p%.jsonnet}"
        done' _ {} + |
      jty -i

Evaluate each .jsonnet file under the current directory,
and for each file foo.jsonnet save a relative yml/foo.yml file
(useful for tools that expect only .yml files in a directory;
e.g. `./apps/foo.jsonnet` -> `./apps/yml/foo.yml`):

    find . -name '*.jsonnet' \
      -exec bash -c 'for p in "$@"; do
        printf "%s\n%s/yml/%s.yml\n" "$p" "$(dirname "$p")" "$(basename "$p" .jsonnet)"
        done' _ {} + |
      jty -i

## Performance

We have one self-contained repository with 22 .jsonnet files that import 17 unique .libsonnet files.
Our old pipeline looked like:

```
time (find . -name '*.jsonnet' -print0 |
      parallel -q -0 bash -c "jsonnet -S -e \"std.manifestYamlStream(import '{}')\" |
      './yamlfmt' > '{.}.yml'")

real	0m10.901s
user	1m21.588s
sys	0m1.252s
```

Where jsonnet is the standard C++ implementation, and `yamlfmt` is a simple Go program that parses YAML and reformats it and emits it to stdout.

If we switch it from the official `jsonnet` binary to the result of `go build github.com/google/go-jsonnet/cmd/jsonnet`, it speeds up quite a bit:

```
real	0m0.951s
user	0m5.799s
sys	0m0.931s
```

If you were to instead drop the yamlfmt command and use [`kubecfg show`](https://github.com/bitnami/kubecfg), you can cut the real time roughly in half:

```
time (find . -name '*.jsonnet' -print0 |
      parallel -q -0 bash -c "kubecfg show "{}" > '{.}.yml'")

real	0m0.448s
user	0m1.990s
sys	0m0.778s
```

But kubecfg offers a lot of functionality that we don't need to just evaluate a lot of Jsonnet.
Note also that kubecfg formats the resulting YAML with slightly different indentation,
although the result is effectively the same.

If we change the pipeline altogether to use jty, the real time drops to about half of the time of using kubecfg,
not to mention the user time being cut to about 18% and the system time cut to about 10%.

```
time (find . -name '*.jsonnet' -exec bash -c 'for p in "$@"; do
        printf "%s\n%s.yml\n" "$p" "${p%.jsonnet}"
        done' _ {} + |
      jty -i)

real	0m0.249s
user	0m0.372s
sys	0m0.080s
```
