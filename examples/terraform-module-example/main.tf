terraform {
  required_providers {
    powerdns = {
      source = "pan-net/powerdns"
    }
  }
}

provider "powerdns" {
  api_key    = "myAPIkey"
  server_url = "http://mydns.domain.tld:8080"
}

resource "powerdns_record" "ip_record" {
  for_each =  { for i, record in local.helper_list : i => record }
  zone    = "subdomain.tld."
  name     = "${each.value.hostname}."
  type     = "A"
  ttl      = 300
  records  = [each.value.ip]
}

resource "awx_inventory" "generated" {
  for_each = var.inventories
  name            = each.key
  description     = each.value.confluence_url
  organisation_id = 1
}

resource "awx_host" "host" {  
  for_each =  { for i, record in local.helper_list : i => record }
  inventory_id = awx_inventory.generated[each.value.inventory].id
  name         = each.value.hostname
  enabled      = true
}