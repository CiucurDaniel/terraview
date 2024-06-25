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

## Current example of a generated diagram 

![Simple diagram](diagram_20240614_172636.png)


# Development 

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

## GOALS

- [x] parseaza graf terraform
- [ ] sterge resursele care nu au sens sau nu fac parte din provider
- [x] adauga label cu imaginea
- [ ] modifica numele resurselor
- [x] print graf in jpeg

## TODOS

Issue with SVGs
https://stackoverflow.com/questions/49819164/graphviz-nodes-of-svg-images-do-not-get-inserted-if-output-is-svg

Separate labels from Image itself
https://stackoverflow.com/questions/58832678/how-to-separate-picture-and-label-of-a-node-with-graphviz

Parse HCL
https://getcoal.medium.com/golang-handling-terraform-files-a37371c621e3

Parse state file
https://github.com/fujiwara/tfstate-lookup/tree/main