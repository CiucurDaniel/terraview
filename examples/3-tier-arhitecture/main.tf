terraform {
  required_version = ">= 1.5.6"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "project" {
  name     = "project-rg"
  location = "West Europe"
}

# VNet with separate subnets for each tier
resource "azurerm_virtual_network" "project_vnet" {
  name                = "project-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
}

# Subnets for each tier
resource "azurerm_subnet" "db_subnet" {
  name                 = "db-subnet"
  resource_group_name  = azurerm_resource_group.project.name
  virtual_network_name = azurerm_virtual_network.project_vnet.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_subnet" "middleware_subnet" {
  name                 = "middleware-subnet"
  resource_group_name  = azurerm_resource_group.project.name
  virtual_network_name = azurerm_virtual_network.project_vnet.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_subnet" "ui_subnet" {
  name                 = "ui-subnet"
  resource_group_name  = azurerm_resource_group.project.name
  virtual_network_name = azurerm_virtual_network.project_vnet.name
  address_prefixes     = ["10.0.3.0/24"]
}

# Database Layer VMs
resource "azurerm_network_interface" "db_nic" {
  count               = 3
  name                = "db-nic-${count.index}"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.db_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_linux_virtual_machine" "db_vm" {
  count                 = 3
  name                  = "db-vm-${count.index}"
  location              = azurerm_resource_group.project.location
  resource_group_name   = azurerm_resource_group.project.name
  network_interface_ids = [azurerm_network_interface.db_nic[count.index].id]
  size                  = "Standard_B1s"

  admin_username        = "adminuser"
  admin_password        = "Password123!"
  disable_password_authentication = false

  os_disk {
    name              = "osdisk-db-${count.index}"
    caching           = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

# Middleware Layer VMs
resource "azurerm_network_interface" "middleware_nic" {
  count               = 3
  name                = "middleware-nic-${count.index}"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.middleware_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_linux_virtual_machine" "middleware_vm" {
  count                 = 3
  name                  = "middleware-vm-${count.index}"
  location              = azurerm_resource_group.project.location
  resource_group_name   = azurerm_resource_group.project.name
  network_interface_ids = [azurerm_network_interface.middleware_nic[count.index].id]
  size                  = "Standard_B1s"

  admin_username        = "adminuser"
  admin_password        = "Password123!"
  disable_password_authentication = false

  os_disk {
    name              = "osdisk-middleware-${count.index}"
    caching           = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

# UI Layer VMs
resource "azurerm_network_interface" "ui_nic" {
  count               = 3
  name                = "ui-nic-${count.index}"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.ui_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_linux_virtual_machine" "ui_vm" {
  count                 = 3
  name                  = "ui-vm-${count.index}"
  location              = azurerm_resource_group.project.location
  resource_group_name   = azurerm_resource_group.project.name
  network_interface_ids = [azurerm_network_interface.ui_nic[count.index].id]
  size                  = "Standard_B1s"

  admin_username        = "adminuser"
  admin_password        = "Password123!"
  disable_password_authentication = false

  os_disk {
    name              = "osdisk-ui-${count.index}"
    caching           = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

# Load Balancers
resource "azurerm_lb" "db_lb" {
  name                = "db-lb"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                          = "db-lb-frontend"
    subnet_id                     = azurerm_subnet.db_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_lb" "middleware_lb" {
  name                = "middleware-lb"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                          = "middleware-lb-frontend"
    subnet_id                     = azurerm_subnet.middleware_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_lb" "ui_lb" {
  name                = "ui-lb"
  location            = azurerm_resource_group.project.location
  resource_group_name = azurerm_resource_group.project.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                          = "ui-lb-frontend"
    subnet_id                     = azurerm_subnet.ui_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}
