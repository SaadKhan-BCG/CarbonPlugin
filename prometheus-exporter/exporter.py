"""Application exporter"""

import datetime
import os
import random
import time
from urllib.parse import quote

import requests
from prometheus_client import start_http_server, Gauge, Enum


class AppMetrics:
    """
    Representation of Prometheus metrics and loop to fetch and transform
    application metrics into Prometheus metrics.
    """
    regions = ['westus', 'eastus', 'uksouth']

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

    # TODO refactor out into separate interface?
    def get_current_carbon_emissions(self, location):
        current_time = datetime.datetime.utcnow()
        url_time = quote(current_time.strftime("%Y-%m-%dT%H:%M"))
        prev_time = current_time - datetime.timedelta(minutes=1)
        prev_url_time = quote(prev_time.strftime("%Y-%m-%dT%H:%M"))

        baseurl = f"http://{self.app_host}:{self.app_port}"
        if self.app_url:
            baseurl = self.app_url
        request_url = f"{baseurl}/emissions/bylocation?location={location}&time={prev_url_time}&toTime={url_time}"

        resp = requests.get(url=request_url)
        return resp.json()[0]['rating']

    def run_metrics_loop(self):
        """Metrics fetching loop"""

        while True:
            self.fetch()
            time.sleep(self.polling_interval_seconds)

    def fetch(self):
        """
        Get metrics from application and refresh Prometheus metrics with
        new values.
        """
        self.health.state("healthy")

        current_power_consumption = random.randint(10, 100)

        self.power_usage.set(current_power_consumption)
        for region in self.regions:
            carbon = self.get_current_carbon_emissions(region)
            self.carbon_consumption.labels(region).set(current_power_consumption * carbon)


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
