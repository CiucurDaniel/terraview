# Terraview

![Build status](https://github.com/CiucurDaniel/landscape/actions/workflows/build-docker-image.yml/badge.svg)

A CLI tool for generating cloud diagrams from Terraform code.
Go from this:

```bash
terraform graph > simple-plan.dot
dot -Tjpeg simple-plan.dot -o diagram.jpg
```

To this:

```bash
terraview print .\terraform_example\ --format png
```

## Current example of generated diagrams 

![Simple diagram](diagram_20240614_172636.png)

![Second Simple diagram](diagram_9517784552.png)

# Development 

Useful commands for development only.

## Cobra-cli

Add new command:

```bash
cobra-cli add print  
```

## Run code

```bash
go run main.go print .\terraform_example\ --format png

go run main.go print .\terraform_example\ 

dot -Tjpeg diagram.dot -o diagram.jpg
```