# landscape

A CLI tool for generating cloud diagrams from Terraform code.

```bash
terraform graph > simple-plan.dot
```

# GOALS

* parseaza graf terraform
* sterge resursele care nu au sens sau nu fac parte din provider
* adauga label cu imaginea
* modifica numele resurselor
* print graf in jpeg

# Cobra-cli

Add new command:

```bash
cobra-cli add print  
```
