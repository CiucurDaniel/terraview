
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

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "East US"
}

# VNet and Subnet for Database Layer
resource "azurerm_virtual_network" "db_vnet" {
  name                = "db-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "db_subnet" {
  name                 = "db-subnet"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.db_vnet.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_network_interface" "db_nic" {
  count               = 3
  name                = "db-nic-${count.index}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.db_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "db_vm" {
  count                 = 3
  name                  = "db-vm-${count.index}"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.db_nic[count.index].id]
  vm_size               = "Standard_DS1_v2"

  storage_os_disk {
    name              = "osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hostname"
    admin_username = "adminuser"
    admin_password = "Password123!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

# VNet and Subnet for Middleware Layer
resource "azurerm_virtual_network" "middleware_vnet" {
  name                = "middleware-vnet"
  address_space       = ["10.1.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "middleware_subnet" {
  name                 = "middleware-subnet"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.middleware_vnet.name
  address_prefixes     = ["10.1.1.0/24"]
}

resource "azurerm_network_interface" "middleware_nic" {
  count               = 3
  name                = "middleware-nic-${count.index}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.middleware_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "middleware_vm" {
  count                 = 3
  name                  = "middleware-vm-${count.index}"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.middleware_nic[count.index].id]
  vm_size               = "Standard_DS1_v2"

  storage_os_disk {
    name              = "osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hostname"
    admin_username = "adminuser"
    admin_password = "Password123!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

# VNet and Subnet for UI Layer
resource "azurerm_virtual_network" "ui_vnet" {
  name                = "ui-vnet"
  address_space       = ["10.2.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "ui_subnet" {
  name                 = "ui-subnet"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.ui_vnet.name
  address_prefixes     = ["10.2.1.0/24"]
}

resource "azurerm_network_interface" "ui_nic" {
  count               = 3
  name                = "ui-nic-${count.index}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.ui_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "ui_vm" {
  count                 = 3
  name                  = "ui-vm-${count.index}"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.ui_nic[count.index].id]
  vm_size               = "Standard_DS1_v2"

  storage_os_disk {
    name              = "osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hostname"
    admin_username = "adminuser"
    admin_password = "Password123!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
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
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                 = "db-lb-frontend"
    subnet_id            = azurerm_subnet.db_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_lb" "middleware_lb" {
  name                = "middleware-lb"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                 = "middleware-lb-frontend"
    subnet_id            = azurerm_subnet.middleware_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_lb" "ui_lb" {
  name                = "ui-lb"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                 = "ui-lb-frontend"
    subnet_id            = azurerm_subnet.ui_subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

# This is a basic example. Further configurations like backend pools, health probes,
# and load balancing rules should be added to complete the setup.
