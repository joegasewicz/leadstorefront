terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.34.1"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

data "digitalocean_ssh_key" "default" {
  name = var.ssh_key_name
}

resource "digitalocean_project" "leadstorefront" {
  name        = "LeadStorefront"
  purpose     = "Web Application"
  environment = "Development"
  description = "Storefront platform for leadstorefront.com"
  is_default  = true
}

resource "digitalocean_droplet" "db" {
  name       = "leadstorefront-db"
  region     = var.region
  size       = "s-1vcpu-1gb"
  image      = "docker-20-04"
  ssh_keys   = [data.digitalocean_ssh_key.default.id]
  monitoring = true
  tags       = ["leadstorefront-db"]

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file(pathexpand(var.pvt_key))
    host        = self.ipv4_address
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait"
    ]
  }
}

resource "digitalocean_droplet" "app" {
  name       = "leadstorefront-app"
  region     = var.region
  size       = "s-1vcpu-1gb"
  image      = "docker-20-04"
  ssh_keys   = [data.digitalocean_ssh_key.default.id]
  monitoring = true
  tags       = ["leadstorefront-app"]

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file(pathexpand(var.pvt_key))
    host        = self.ipv4_address
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait"
    ]
  }
}

resource "digitalocean_volume" "caddy_data" {
  name                    = "leadstorefront-caddy-data"
  region                  = var.region
  size                    = 1
  initial_filesystem_type = "ext4"
  tags                    = ["caddy", "leadstorefront"]
}

resource "digitalocean_volume_attachment" "caddy_attach" {
  droplet_id = digitalocean_droplet.app.id
  volume_id  = digitalocean_volume.caddy_data.id
}

resource "digitalocean_project_resources" "attach" {
  project = digitalocean_project.leadstorefront.id
  resources = [
    digitalocean_droplet.db.urn,
    digitalocean_droplet.app.urn,
    digitalocean_volume.caddy_data.urn
  ]
}

resource "digitalocean_floating_ip" "app_floating_ip" {
  region = digitalocean_droplet.app.region
}

resource "digitalocean_floating_ip" "db_floating_ip" {
  region = digitalocean_droplet.db.region
}

resource "digitalocean_floating_ip_assignment" "assign_app_ip" {
  ip_address = digitalocean_floating_ip.app_floating_ip.ip_address
  droplet_id = digitalocean_droplet.app.id

  depends_on = [digitalocean_droplet.app]
}

resource "digitalocean_floating_ip_assignment" "assign_db_ip" {
  ip_address = digitalocean_floating_ip.db_floating_ip.ip_address
  droplet_id = digitalocean_droplet.db.id

  depends_on = [digitalocean_droplet.db]
}

output "app_floating_ip" {
  value = digitalocean_floating_ip.app_floating_ip.ip_address
}

output "db_floating_ip" {
  value = digitalocean_floating_ip.db_floating_ip.ip_address
}

output "app_ip" {
  value = digitalocean_floating_ip.app_floating_ip.ip_address
}

output "db_ip" {
  value = digitalocean_floating_ip.db_floating_ip.ip_address
}

resource "digitalocean_firewall" "leadstorefront_app_firewall" {
  name        = "leadstorefront-app-firewall"
  droplet_ids = [digitalocean_droplet.app.id]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = [var.my_ip, "0.0.0.0/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "8002"
    source_addresses = [var.my_ip]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

resource "digitalocean_firewall" "leadstorefront_db_firewall" {
  name        = "leadstorefront-db-firewall"
  droplet_ids = [digitalocean_droplet.db.id]

  inbound_rule {
    protocol   = "tcp"
    port_range = "22"
    source_addresses = [
      var.my_ip,
      digitalocean_droplet.app.ipv4_address_private,
    ]
  }

  inbound_rule {
    protocol           = "tcp"
    port_range         = "5432"
    source_droplet_ids = [digitalocean_droplet.app.id]
    source_addresses = [
      var.my_ip,
      digitalocean_droplet.app.ipv4_address_private,
    ]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}
