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


class AppMetrics:
    """
    Representation of Prometheus metrics and loop to fetch and transform
    application metrics into Prometheus metrics.
    """

    def __init__(self, polling_interval_seconds, app_host, app_port, app_url=None):
        self.app_host = app_host
        self.app_port = app_port
        self.app_url = app_url
        self.polling_interval_seconds = polling_interval_seconds

        # Prometheus metrics to collect
        self.health = Enum("app_health", "Health", states=["healthy", "unhealthy"])
        self.power_usage = Gauge("power_usage", "Power Useage (Watts)")
        self.carbon_consumption = Gauge(name="carbon_consumption",
                                        documentation="Carbon Consumed (gCo2Eq)",
                                        labelnames=["region"])
        self.carbon_consumption_time = Gauge(name="carbon_consumption_time",
                                             documentation="Carbon Consumed at Given Hour(gCo2Eq)",
                                             labelnames=["region", "time"])

    # TODO refactor out carbon methods into separate class
    def __get_carbon_emissions(self, location, prev_time, to_time):
        """

        :param location: Location to extract Carbon Data from
        :param prev_time: url encoded start time
        :param to_time: url encoded end time
        :return: Carbon consumption at given time and region in gCO2eq/watt
        """
        baseurl = f"http://{self.app_host}:{self.app_port}"
        if self.app_url:
            baseurl = self.app_url

        request_url = f"{baseurl}/emissions/bylocation?location={location}&time={prev_time}&toTime={to_time}"

        resp = requests.get(url=request_url)
        return resp.json()[0]['rating']

    def __get_carbon_emissions_utc(self, location, utc_time):
        """

        :param location: Location to extract Carbon Data from
        :param utc_time: python date object in utc time you want the carbon data at
        :return: Carbon consumption at given time and region in gCO2eq/watt
        """
        url_time = quote(utc_time.strftime("%Y-%m-%dT%H:%M"))
        prev_time = utc_time - datetime.timedelta(minutes=1)
        prev_url_time = quote(prev_time.strftime("%Y-%m-%dT%H:%M"))

        return self.__get_carbon_emissions(location, prev_url_time, url_time)

    def __get_current_carbon_emissions(self, location):
        current_time = datetime.datetime.utcnow()
        return self.__get_carbon_emissions_utc(location, current_time)

    def __set_carbon_per_date_time(self, location, current_power_consumption):
        start_time = datetime.datetime.utcnow() - datetime.timedelta(days=1)
        start_time = start_time - datetime.timedelta(hours=datetime.datetime.now().hour)

        for i in range(0, 23):
            cur_time = start_time + datetime.timedelta(hours=i)
            carbon = self.__get_carbon_emissions_utc(location, cur_time)
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
                carbon = self.__get_current_carbon_emissions(region)
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
