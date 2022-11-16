# Carbon Plugin

## Background
This tool is designed to help developers be Carbon Aware by monitoring the carbon consumption of their applications and showing what it would be like in different times of day/regions.
Developers can then use this data to shift expensive tasks to regions and times were their consumption would be minimised.

The Carbon Plugin works by gathering the resource use of docker containers, and combining this with the Carbon Aware SDK by the Green Software Foundation
to produce a single dashboard plotting the total carbon of the app if it were ran at various countries and times.


## Requirements
- docker


## Quick Start
- Clone this repo or just use the compose file (also included as compose.yml in the repo):

 ```
 version: "3.7"
services:
  prometheus:
    image: saadbcg/carbon-plugin-prometheus
    ports:
      - 9090:9090
  grafana:
    image: saadbcg/carbon-plugin-grafana
    ports:
      - 3001:3001
    links:
      - prometheus
  prometheus-exporter:
    image: saadbcg/carbon-plugin-prometheus-exporter
    environment:
      TIME_REGIONS: "uksouth, westus"
      CARBON_SDK_URL: "https://carbon-aware-api.azurewebsites.net"
    ports:
      - 9877:9877
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    links:
      - carbon-aware-sdk-webapi
  carbon-aware-sdk-webapi:
    image: saadbcg/carbon-aware-sdk-webapi  # Note this is built manually from the carbon aware source code
    environment:
      CarbonAwareVars__CarbonIntensityDataSource: WattTime
      WattTimeClient__Username: ${WattTimeClient__Username}
      WattTimeClient__Password: ${WattTimeClient__Password}
    ports:
      - 8080:80
 ```
- create a .env file in the same directory with the following:
```
COMPOSE_PROJECT_NAME=carbon-plugin
```
- run ``docker compose up``
- Open up the dashboard at http://localhost:3001 (note chrome often has issues with![how-to](https://user-images.githubusercontent.com/101206684/202210155-212f90e6-70fa-47de-9213-8c8cd0e20af8.gif)
 http sites, recommend using firefox/safari if you are having issues)
  - credentials for Grafana are simply:
    - **username: admin**
    - **password: admin**
- Select the "Carbon" dashboard to see carbon data
- Select your region using the drop-down at the top, you can also filter by Container Name using the other drop down
