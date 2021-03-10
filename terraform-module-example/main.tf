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
  for_each = var.records
  zone    = "subdomain.tld."
  name    = each.value
  type    = "A"
  ttl     = 300
  records = [each.key]
}