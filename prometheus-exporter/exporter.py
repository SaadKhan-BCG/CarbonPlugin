"""Application exporter"""

import datetime
import logging
import os
import time

from prometheus_client import start_http_server, Gauge, Enum

from carbon_emissions import CarbonEmissions
from regions import regions
from container_stats import get_stats


class AppMetrics:
    """
    Representation of Prometheus metrics and loop to fetch and transform
    application metrics into Prometheus metrics.
    """

    def __init__(self, polling_interval_seconds, app_host, app_port, app_url=None):
        self.polling_interval_seconds = polling_interval_seconds
        self.carbon_emissions_client = CarbonEmissions(app_host, app_port, app_url)

        # Prometheus metrics to collect
        self.health = Enum("app_health", "Health", states=["healthy", "unhealthy"])
        self.power_usage = Gauge("power_usage", "Power Useage (Watts)", labelnames=["container_name"])
        self.carbon_consumption = Gauge(name="carbon_consumption",
                                        documentation="Carbon Consumed (gCo2Eq)",
                                        labelnames=["region", "container_name"])
        self.carbon_consumption_time = Gauge(name="carbon_consumption_time",
                                             documentation="Carbon Consumed at Given Hour(gCo2Eq)",
                                             labelnames=["region", "time", "container_name"])

    def __set_carbon_per_date_time(self, location, current_power_consumption, container):
        """
        Get carbon Metrics for given location throughout the day (from midnight to midnight the previous day)
        Useful for identifying optimal times to run workloads

        :param location: Region workload is ran
        :param current_power_consumption: Power consumption in watts of your application/process
        """
        start_time = datetime.datetime.utcnow() - datetime.timedelta(days=1)
        start_time = start_time - datetime.timedelta(hours=datetime.datetime.now().hour)

        for i in range(0, 23):
            cur_time = start_time + datetime.timedelta(hours=i)
            carbon = self.carbon_emissions_client.get_carbon_emissions_utc(location, cur_time)
            self.carbon_consumption_time.labels(location, i, container).set(current_power_consumption * carbon)
            print(f"container: {container} region: {location} carbon: {carbon} power: {current_power_consumption}")

    def __fetch(self):
        """
        Get metrics from application and refresh Prometheus metrics with
        new values.
        """
        self.health.state("healthy")
        print("FETCHING DATA")
        current_power_consumption = get_stats()
        for container, power in current_power_consumption.items():

            self.power_usage.labels(container).set(power)
            print("*****************************************")
            print(f"Power for Container: {container} Watts: {power}")

            for region in regions:
                try:
                    carbon = self.carbon_emissions_client.get_current_carbon_emissions(region)
                    self.carbon_consumption.labels(region, container).set(power * carbon)
                    # self.__set_carbon_per_date_time(region, power, container) # TODO fix this, its currently taking too long and causing things to break
                except Exception as err:  # TODO make this better exception handling
                    print(f"Failed for region: {region} due to exception: \n{err}")
                    logging.warning(f"Failed for region: {region} due to exception: \n{err}")
        print("*****************************************")

    def run_metrics_loop(self):
        """Metrics fetching loop"""

        while True:
            self.__fetch()
            time.sleep(self.polling_interval_seconds)


def main():
    """Main entry point"""
    logging.info("STARTING EXPORTER")
    polling_interval_seconds = int(os.getenv("POLLING_INTERVAL_SECONDS", "10"))
    app_port = int(os.getenv("CARBON_SDK_PORT", "80"))
    app_host = str(os.getenv("CARBON_SDK_HOST", "carbon-aware-sdk-webapi"))
    app_url = os.getenv("CARBON_SDK_URL")
    exporter_port = int(os.getenv("EXPORTER_PORT", "9877"))

    app_metrics = AppMetrics(
        polling_interval_seconds=polling_interval_seconds,
        app_host=app_host,
        app_port=app_port,
        app_url=app_url
    )
    start_http_server(exporter_port)
    app_metrics.run_metrics_loop()


def test():
    print("STARTING EXPORTER")
    polling_interval_seconds = int(os.getenv("POLLING_INTERVAL_SECONDS", "10"))
    app_port = int(os.getenv("CARBON_SDK_PORT", "80"))
    app_host = str(os.getenv("CARBON_SDK_HOST", "carbon-aware-sdk-webapi"))
    app_url = os.getenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
    exporter_port = int(os.getenv("EXPORTER_PORT", "9877"))

    app_metrics = AppMetrics(
        polling_interval_seconds=polling_interval_seconds,
        app_host=app_host,
        app_port=app_port,
        app_url=app_url
    )
    app_metrics.run_metrics_loop()


if __name__ == "__main__":
    main()
