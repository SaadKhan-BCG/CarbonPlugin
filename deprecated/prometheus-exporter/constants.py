# TODO make this better
# Im not sure if there are more available regions than this,
# once we have a wattime account we should be able to look them up with:
# https://www.watttime.org/api-documentation/#list-of-grid-regions
regions = [
    "australiacentral",
    "australiacentral2",
    "australiaeast",
    "australiasoutheast",
    "canadacentral",
    "canadaeast",
    "centralus",
    "centraluseuap",
    "eastus",
    "eastus2",
    "eastus2euap",
    "northcentralus",
    "northeurope",
    "southcentralus",
    "uksouth",
    "ukwest",
    "westcentralus",
    "westus",
    "westus2",
    "westus3"]

energy_profile = {
    'CPU': 45,
    'MEM': 10/128,
    'DISK': 1.52/1000,
    'NETWORK': 11,
    'PUE': 1.4
}

# Name of the docker project, used to filter out the plugin's own containers from stat collection
project_name = "carbon-plugin"
