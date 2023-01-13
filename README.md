# Carbon Plugin

## Background
This tool is designed to help developers be Carbon Aware by monitoring the carbon consumption of their applications and showing what it would be like in different times of day/regions.
Developers can then use this data to shift expensive tasks to regions and times were their consumption would be minimised.

The Carbon Plugin works by gathering the resource use of docker containers, and combining this with the Carbon Aware SDK by the Green Software Foundation
to produce a single dashboard plotting the total carbon of the app if it were ran at various countries and times.

![image](https://user-images.githubusercontent.com/101206684/203093429-4ce892c2-1bd9-49e8-a13a-f713e10248f4.png)

## Requirements
- docker


## Quick Start

### Option 1: (If you dont need to configure anything)
Just run 
```curl https://raw.githubusercontent.com/SaadKhan-BCG/CarbonPlugin/main/compose.yml | docker compose -f -  -p carbon-plugin up```

And check 4. in Option 2 for details on the dashboard access

### Option 2: (allows you to configure env variables etc yourself)
1. Clone this repo or just use the compose file (also included as compose.yml in the repo):

 ```
version: "3.7"
services:
  prometheus:
    image: saadbcg/carbon-plugin-prometheus:2.0.0
    ports:
      - 9090:9090
  grafana:
    image: saadbcg/carbon-plugin-grafana:1.0.0
    ports:
      - 3001:3001
    links:
      - prometheus
  prometheus-exporter:
    image: saadbcg/carbon-plugin-prometheus-exporter:2.0.0
    environment:
      TIME_REGIONS: "westus,uksouth"
      CARBON_SDK_URL: "https://carbon-aware-api.azurewebsites.net"
#      CARBON_SDK_PORT: 80
#      CARBON_SDK_HOST: carbon-aware-sdk-webapi
    ports:
      - 2112:2112
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    links:
      - carbon-aware-sdk-webapi
  carbon-aware-sdk-webapi:
    image: saadbcg/carbon-aware-sdk-webapi:0.0.1  # Note this is built manually from the carbon aware source code
    environment:
      CarbonAwareVars__CarbonIntensityDataSource: WattTime
    ports:
      - 8080:80
 ```
2. create a .env file in the same directory with the following:
```
COMPOSE_PROJECT_NAME=carbon-plugin
```
3. run ``docker compose up``
4. Open up the dashboard at http://localhost:3001 (note chrome often has issues with
 http sites, recommend using firefox/safari if you are having issues)
    - credentials for Grafana are simply:
      - **username: admin**
      - **password: admin**
    - Select the "Carbon" dashboard to see carbon data
    - Select your region using the drop-down at the top, you can also filter by Container Name using the other drop down



https://user-images.githubusercontent.com/101206684/202211297-7bda7783-8b10-4401-b14c-98d9a525e48f.mov


## Time Regions
The plugin is designed to provide data on different regions around the world and times of day.
However due to performance reasons we do not expport every combination of time/region
If you wish to see the carbon impact of your app in a particular region at different times of day you must
set it in the ``TIME_REGIONS`` env variable on the prometheus-exporter (see sample with uksouth and westus in this README/compose.yml)

For all regions provided, you will be able to compare your app performance if it were ran at different times of day in 4 hour intervals (ie at 4am, 8am etc) via the ``Total Carbon Consumed By Time`` panel

## Common Issues

- Exporter is not returning any data
  - Its possible the docker volumes are full (docker does not clean up by itself)
  - Solution: 
    - first try ``docker volume prune -f`` and  ``docker system prune -f``
    - If needed also restart docker after these
- Exporter is working, but I only see some of my containers in grafana
  - Grafana is often a little slow to grab all metrics so often waiting a few minutes and refreshing will solve it
  - If you wish to force it to refresh: click settings (cog next to time range at the top of dashboard) -> Variables -> (select variable you wish to refresh, usually container_name) -> Run query

## Release Process
This app is built and deployed using Github Actions, the release process is therefore fairly simple:
- Merge code to main
- Generate a Release on Github (create a new tag from main with your new version number of choice) and let CI build it

## Contributing
If you wish to contribute to this project please make a relevent PR and contact Saad Khan  (khan.saad@bcg.com) for review.
For the convenience of developers, there is a build-compose.yml included in this repo. This will rebuild all local containers to allow you to rapidly test changes.
Using this is fairly simple:
- Make local code changes (to the exporter, config for prometheus etc)
- Run ``docker compose -f build-compose.yml up --build`` to see changes applied in dashboard
- Alternatively to test the prometheus-exporter (in particular its low power mode options), you can try out different running modes using the carbon-monitor/example application
    
## Carbon Calculation Methodology
TLDR, we pull carbon data from the GSF carbon-awaresdk and multiple by power consumption (estimated using docker stats) to get overall carbon consumption over time.


### Carbon
We are relying on the Green Software Foundation's Carbon Aware SDK https://github.com/Green-Software-Foundation/carbon-aware-sdk
for all carbon data. This sdk takes a location and time period as input and provides a carbon metric in gCo2Eq/kwh.
We can query the current time in different regions to get live data, and yesterday's data throughout the day to get estimates for running your app at various times


A possible line of improvement here would be to take multiple metrics and average them to get a more accurate estimate per time of day

### Power
Gathering accurate power consumption data is tricky and very platform/OS specific. As a result we rely on an estimate of power consumption relying on the methodology published in GreenFrame https://github.com/marmelab/greenframe-cli/blob/main/src/model/README.md.

To do this we gather current cpu, memory and network utilisation stats from the docker stats and convert these to power numbers using the formula given in GreenFrame.


Note this formula is an **Estimate**, not a true measure of power. However, the true power consumption is only a scaling factor off (depending on your hardware/OS) and therefore relying on this estimate does not in any way affect the functionality of this tool as a means to compare regions, times to improve carbon consumption.

#### Possible Improvements

We could improve the accuracy of power collection by having multiple running modes, defaulting back to the estimator if a more accurate metric cannot be found
- for Linux systems there is a tool called scaphandre https://github.com/hubblo-org/scaphandre which provides excellent metrics on docker container power
  - Excellent solution, even includes a prometheus exporter itself so would integrate nicely with this plugin
  - Includes kubernetes support
  - Does not currently support any cloud provider (requires they implement a hypervisor that at time of writing AWS, GCP and Azure dont support)
- Implement a Kubernetes scraper
  - Would work very similarly to the existing docker solution but scrape kubernetes pods instead
- Mac
  - Apple does export some accurate power information via a tool called powermetrics. However, it's behaviour varies for m1 vs intel macs so a solution would need to be developed for both separately 
