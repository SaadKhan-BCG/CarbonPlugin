"""Application exporter"""

import datetime
import logging
import os
import time
import asyncio

from prometheus_client import start_http_server, Gauge, Enum

from carbon_emissions import CarbonEmissions
from constants import regions
from container_stats import get_stats


def background(f):
    def wrapped(*args, **kwargs):
        return asyncio.get_event_loop().run_in_executor(None, f, *args, **kwargs)

    return wrapped


class AppMetrics:
    """
    Representation of Prometheus metrics and loop to fetch and transform
    application metrics into Prometheus metrics.
    """

    def __init__(self, polling_interval_seconds, app_host, app_port, time_regions, app_url=None):
        self.polling_interval_seconds = polling_interval_seconds
        self.time_regions = time_regions
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

        for i in range(0, 6):
            time_displacement = i * 4
            cur_time = start_time + datetime.timedelta(hours=time_displacement)
            carbon = self.carbon_emissions_client.get_carbon_emissions_utc(location, cur_time)
            self.carbon_consumption_time.labels(location, time_displacement, container).set(
                current_power_consumption * carbon)
            logging.debug(
                f"container: {container} region: {location} carbon: {carbon} power: {current_power_consumption}")

    @background
    def __set_carbon_per_container(self, container, power):
        self.power_usage.labels(container).set(power)
        logging.info(f"Power for Container: {container} Watts: {power}")

        for region in regions:
            try:
                carbon = self.carbon_emissions_client.get_current_carbon_emissions(region)
                self.carbon_consumption.labels(region, container).set(power * carbon)
            except Exception as err:  # TODO make this better exception handling
                logging.info(f"Failed for region: {region} due to exception: \n{err}")
                logging.warning(f"Failed for region: {region} due to exception: \n{err}")

        if self.time_regions:
            for region in self.time_regions:
                try:
                    self.__set_carbon_per_date_time(region, power, container)
                except Exception as err:  # TODO make this better exception handling
                    logging.info(f"Failed for region: {region} due to exception: \n{err}")
                    logging.warning(f"Failed for region: {region} due to exception: \n{err}")

    def __fetch(self):
        """
        Get metrics from application and refresh Prometheus metrics with
        new values.
        """
        self.health.state("healthy")
        logging.info("FETCHING DATA")
        current_power_consumption = get_stats()
        logging.info("*****************************************")
        for container, power in current_power_consumption.items():
            self.__set_carbon_per_container(container, power)
        logging.info("*****************************************")

    def run_metrics_loop(self):
        """Metrics fetching loop"""

        while True:
            self.__fetch()
            time.sleep(self.polling_interval_seconds)

    def run_timed_test(self):
        logging.info("*****************************************")
        logging.info("*****************************************")
        start = datetime.datetime.utcnow()
        logging.info(f"START TIME: {start}")

        for i in range(0, 10):
            self.__fetch()
            end = datetime.datetime.utcnow()
            logging.info(f"End of run {i} Time: {end}")
            logging.info(f"End of run {i} Time Taken: {end - start}")

        end = datetime.datetime.utcnow()
        logging.info(f"END TIME: {end}")
        logging.info(f"TOTAL TIME: {end - start}")
        # INFO:root:TOTAL TIME: 0:03:21.228341
        logging.info("*****************************************")
        logging.info("*****************************************")


def main():
    """Main entry point"""
    logging.info("STARTING EXPORTER")
    polling_interval_seconds = int(os.getenv("POLLING_INTERVAL_SECONDS", "5"))
    app_port = int(os.getenv("CARBON_SDK_PORT", "80"))
    app_host = str(os.getenv("CARBON_SDK_HOST", "carbon-aware-sdk-webapi"))
    app_url = os.getenv("CARBON_SDK_URL")
    exporter_port = int(os.getenv("EXPORTER_PORT", "9877"))
    time_regions = (os.getenv("TIME_REGIONS")).split(", ")

    app_metrics = AppMetrics(
        polling_interval_seconds=polling_interval_seconds,
        app_host=app_host,
        app_port=app_port,
        time_regions=time_regions,
        app_url=app_url
    )
    start_http_server(exporter_port)
    app_metrics.run_metrics_loop()


def debug():
    """entry point for debugging
        Convenience method to debug the exporter locally
    """
    logging.basicConfig(level=logging.INFO)
    logging.info("STARTING EXPORTER")
    polling_interval_seconds = int(os.getenv("POLLING_INTERVAL_SECONDS", "5"))
    app_port = int(os.getenv("CARBON_SDK_PORT", "80"))
    app_host = str(os.getenv("CARBON_SDK_HOST", "carbon-aware-sdk-webapi"))
    app_url = os.getenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
    time_regions = (os.getenv("TIME_REGIONS")).split(", ")

    app_metrics = AppMetrics(
        polling_interval_seconds=polling_interval_seconds,
        app_host=app_host,
        app_port=app_port,
        time_regions=time_regions,
        app_url=app_url
    )
    app_metrics.run_timed_test()


if __name__ == "__main__":
    main()
    # debug()
