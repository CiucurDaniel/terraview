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

resource "azurerm_resource_group" "online_shop_prod" {
  name     = "online_shop_prod"
  location = "West Europe"
}

resource "azurerm_virtual_network" "online_shop_network_prod" {
  name                = "online_shop_network_prod"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.online_shop_prod.location
  resource_group_name = azurerm_resource_group.online_shop_prod.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.online_shop_prod.name
  virtual_network_name = azurerm_virtual_network.online_shop_network_prod.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "server_public_ip" {
  name                = "server_public_ip"
  location            = azurerm_resource_group.online_shop_prod.location
  resource_group_name = azurerm_resource_group.online_shop_prod.name
  allocation_method   = "Dynamic"
}

resource "azurerm_network_interface" "application_server_nic" {
  name                = "application_server_nic"
  location            = azurerm_resource_group.online_shop_prod.location
  resource_group_name = azurerm_resource_group.online_shop_prod.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.server_public_ip.id
  }
}

resource "azurerm_linux_virtual_machine" "application_server" {
  name                  = "application_server"
  location              = azurerm_resource_group.online_shop_prod.location
  resource_group_name   = azurerm_resource_group.online_shop_prod.name
  network_interface_ids = [azurerm_network_interface.application_server_nic.id]
  size                  = "Standard_B1s"

  admin_username = "adminuser"
  admin_password = "Password1234!"
  disable_password_authentication = false

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  computer_name = "application-server"
}

resource "azurerm_postgresql_server" "orders_db_prod" {
  name                = "orders-db-prod"
  location            = azurerm_resource_group.online_shop_prod.location
  resource_group_name = azurerm_resource_group.online_shop_prod.name

  administrator_login          = "psqladminun"
  administrator_login_password = "H@Sh1CoR3!"

  sku_name   = "B_Gen5_1"
  storage_mb = 5120
  version    = 10

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  auto_grow_enabled            = true
  ssl_enforcement_enabled      = true
}

