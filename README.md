# landscape

A CLI tool for generating cloud diagrams from Terraform code.

```bash
terraform graph > simple-plan.dot
```

# GOALS

- [x] parseaza graf terraform
- [ ] sterge resursele care nu au sens sau nu fac parte din provider
- [x] adauga label cu imaginea
- [ ] modifica numele resurselor
- [ ] print graf in jpeg

# Cobra-cli

Add new command:

```bash
cobra-cli add print  
```

# Run code

```bash
go run main.go print .\terraform_example\ 

dot -Tjpeg diagram.dot -o diagram.jpg
```


# TODOS

Issue with SVGs
https://stackoverflow.com/questions/49819164/graphviz-nodes-of-svg-images-do-not-get-inserted-if-output-is-svg

Separate labels from Image itself
https://stackoverflow.com/questions/58832678/how-to-separate-picture-and-label-of-a-node-with-graphviz

Parse HCL
https://getcoal.medium.com/golang-handling-terraform-files-a37371c621e3