variable "awx_inventories" {
  type = map(object({
    docs_url  = string
    support_ticket  = string
    hosts = map(string) 
  }))
}
// https://stackoverflow.com/questions/63500554/terraform-iterate-over-nested-map?noredirect=1&lq=1
// https://www.terraform.io/docs/language/functions/flatten.html
locals {
  helper_list = flatten([for inventory, items in var.inventories:
                 flatten([for host, ip in items.hosts:
                   {
                     "inventory" = inventory
                     "hostname" = host
                     "ip" = ip
                   }
                 ])
                ])
}