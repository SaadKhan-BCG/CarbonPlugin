"""Application exporter"""

import datetime
import os
import random
import time
import logging
from urllib.parse import quote

import requests
from prometheus_client import start_http_server, Gauge, Enum

from regions import regions

from carbon_emissions import CarbonEmissions


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
        self.power_usage = Gauge("power_usage", "Power Useage (Watts)")
        self.carbon_consumption = Gauge(name="carbon_consumption",
                                        documentation="Carbon Consumed (gCo2Eq)",
                                        labelnames=["region"])
        self.carbon_consumption_time = Gauge(name="carbon_consumption_time",
                                             documentation="Carbon Consumed at Given Hour(gCo2Eq)",
                                             labelnames=["region", "time"])

    def __set_carbon_per_date_time(self, location, current_power_consumption):
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
            self.carbon_consumption_time.labels(location, i).set(current_power_consumption * carbon)

    def __fetch(self):
        """
        Get metrics from application and refresh Prometheus metrics with
        new values.
        """
        self.health.state("healthy")

        current_power_consumption = random.randint(10, 100)

        self.power_usage.set(current_power_consumption)
        for region in regions:
            try:
                carbon = self.carbon_emissions_client.get_current_carbon_emissions(region)
                self.carbon_consumption.labels(region).set(current_power_consumption * carbon)
                self.__set_carbon_per_date_time(region, current_power_consumption)
            except Exception as err:  # TODO make this better exception handling
                logging.warning(f"Failed for region: {region} due to exception: \n{err}")

    def run_metrics_loop(self):
        """Metrics fetching loop"""

        while True:
            self.__fetch()
            time.sleep(self.polling_interval_seconds)


def main():
    """Main entry point"""

    polling_interval_seconds = int(os.getenv("POLLING_INTERVAL_SECONDS", "5"))
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


if __name__ == "__main__":
    main()
