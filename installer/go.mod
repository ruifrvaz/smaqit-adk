module github.com/daisy-tribe/smaqit-adk/installer

go 1.24

require github.com/daisy-tribe/smaqit-adk/src v0.0.0

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace github.com/daisy-tribe/smaqit-adk/src => ../src
