# jty: Jsonnet To Yaml

jty is a simple utility that does one thing: it processes many Jsonnet files and emits YAML to specified output files.

We have a fair sized configuration repository of Jsonnet, and we used to have a multiple-step pipeline to generate the YAML.
It didn't take _long_ to do so, but the runtime was noticeable, especially when you needed to make many small changes in a row and regenerate the YAML each step along the way.

`jty` simply accepts input as pairs of /path/to/input.jsonnet and /path/to/output.yml, and in a single process evaluates all the input Jsonnet to generate the corresponding output YAML.
This way, common .libsonnet files are read and evaluated only once.

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

And if we change the pipeline altogether to use jty, the real time drops to about 25% of the previous step,
not to mention the user time being cut to about 6% of the previous step and the system time cut to about 9%.

```
time (find . -name '*.jsonnet' -exec bash -c 'for p in "$@"; do
        printf "%s\n%s.yml\n" "$p" "${p%.jsonnet}"
        done' _ {} + |
      jty -i)

real	0m0.249s
user	0m0.372s
sys	0m0.080s
```
