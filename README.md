# table-to-terrawx

## About

Web scraper that makes data from HTML table available on Terraform and AWX
## Motivation
Using documentation as a source for creating Terrafrom files and AWX inventories. Let us consider an example where you are given a task to provision a couple of virtual machines using information provided in documentation (ie. Confluence/Notion) in the form of table:
| Hostname      | vCPU | RAM | IP address
| ----------- | ----------- | ----------- | ----------- |
| web-001.subdomain.tld      | 4       |4GB| 10.0.0.2
| lb-001.subdomain.tld    | 2        |2GB|10.0.0.3| 

To provision those VMs, you would need to manualy create Terraform file and AWX inverntory and retype those values. **table-to-terrawx** enables you to scrape the documentation page and extract values in the form of key:value (ie: Hostname:IP) which will be used by other tools in provision process.
## Example
Let's consider example from above:
| Hostname      | vCPU | RAM | IP address
| ----------- | ----------- | ----------- | ----------- |
| web-001.subdomain.tld      | 4       |4GB| 10.0.0.2
| lb-001.subdomain.tld    | 2        |2GB|10.0.0.3 |

We can run the following command to extract values:

    go run main.go \
    --url="https://confluence.domain.tld/display/Site/Page-With-VMs" \
    --username="user"
    --password="password"
    --table-headers="Hostname,vCPU"
    --kv-list="Hostname:IP address"
    
    Output:
    {
      "web-001.subdomain.tld": "10.0.0.2",
      "lb-001.subdomain.tld": "10.0.0.3"
    }
    
## Usage
CLI flags that we can use:

 - **--url** : **(mandatory)** defines URL that will be scraped
 - **--username** : (optional) username if webpage uses authentication
 - **--password**: (optional) password if webpage uses authentication
 - **--table-headers**: **(mandatory)** CSV list of table headers that will be used to match tables on webpage. If table headers do not match, scraped table will be ommited from output
 - **--kv-list**: **(mandatory)** defines table header that will be used as key and table header that will be used as value. Usefull if we want to query for total RAM usage or vCPU count