# table-to-terrawx

## About

Web scraper that makes data from HTML table on Atlassian Confluence available on Terraform and AWX. 
### Note
Main branch focues on geting data from Atlassian Confluence, while general-uri-scrape contains more broad example for getting table data into JSON.
## Motivation
Using documentation as a source for creating Terrafrom files and AWX inventories. Let us consider an example where you are given a task to provision a couple of virtual machines using information provided in documentation (ie. Confluence) in the form of table:
| Hostname      | vCPU | RAM | IP address |
| ----------- | ----------- | ----------- | ----------- |
| web-001.subdomain.tld      | 4       |4GB| 10.0.0.2 |
| lb-001.subdomain.tld    | 2        |2GB|10.0.0.3| 

To provision those VMs, you would need to manualy create Terraform file and AWX inverntory and retype those values. **table-to-terrawx** enables you to scrape the documentation page and extract values in the form of key:value (ie: Hostname:IP) which will be used by other tools in provision process. It enables output to tfvars.json in the format that best suits my needs, but it can be modified very easily.
## Example
Let's consider example from above:
| Hostname      | vCPU | RAM | IP |
| ----------- | ----------- | ----------- | ----------- |
| web-001.subdomain.tld      | 4       |4GB| 10.0.0.2 |
| lb-001.subdomain.tld    | 2        |2GB|10.0.0.3 |

We can run the following command to extract values:

    go run main.go \
    -confl-domain="<subdomain>.atlassian.net" \
    -confl-pageid=1234567 \
    -confl-user="user@domain.tld" \
    -confl-apikey="<generated-api-key-for-account>" \
    -table-headers="Hostname" \
    -kv-list="Hostname:IP" 
    
    Output:
    {
      "web-001.subdomain.tld": "10.0.0.2",
      "lb-001.subdomain.tld": "10.0.0.3"
    }
    
## Usage
CLI flags (all are **mandatory**):

 - **--confl-domain**  : defines fqdn of Atlassian confluence page  
 - **--confl-pageid**  : defines page-id that will be scraped. Can be found inside URL after /pages/ location
 - **--confl-user**    : username for Atlassian cloud
 - **--confl-apikey**  : API key generated for Atlassian cloud. Since confluence does not allow auth with password, API key needs to be generated
 - **--table-headers** : CSV list of table headers that will be used to match tables on webpage. If table headers do not match, scraped table will be ommited from output
 - **--kv-list**       : defines table header that will be used as key and table header that will be used as value. Usefull if we want to query for total RAM usage or vCPU count