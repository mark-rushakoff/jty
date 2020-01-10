# jty: Jsonnet To Yaml

jty is a simple utility that processes many Jsonnet files and emits corresponding YAML files.

We have a fair sized configuration repository of Jsonnet, and we used to have a multiple-step pipeline to generate the YAML.
It didn't take _long_ to do so, but the runtime was noticeable, especially when you needed to make many small changes in a row and regenerate the YAML each step along the way.

`jty` simply accepts input as pairs of /path/to/input.jsonnet and /path/to/output.yml, and in a single process evaluates all the input Jsonnet to generate the corresponding output YAML.
This way, common .libsonnet files are read and evaluated only once.
