"""Application exporter"""

import os
import random
import time
from prometheus_client import start_http_server, Gauge, Enum
import requests
import datetime
from urllib.parse import urlparse, urlencode, quote


class AppMetrics:
    """
    Representation of Prometheus metrics and loop to fetch and transform
    application metrics into Prometheus metrics.
    """
    regions = ['westus', 'eastus', 'uksouth']

    def __init__(self, app_host, app_port, polling_interval_seconds):
        self.app_host = app_host
        self.app_port = app_port
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

        # temporary
        # request_url = f"http://{self.app_host}:{self.app_port}/emissions/bylocation?location={location}&time={prev_url_time}&toTime={url_time}"
        request_url = f"https://{self.app_host}/emissions/bylocation?location={location}&time={prev_url_time}&toTime={url_time}"

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
    app_host = int(os.getenv("CARBON_SDK_HOST", "80"))
    exporter_port = int(os.getenv("EXPORTER_PORT", "9877"))

    app_host = "carbon-aware-api.azurewebsites.net"     # temporary

    app_metrics = AppMetrics(
        app_host=app_host,
        app_port=app_port,
        polling_interval_seconds=polling_interval_seconds
    )
    start_http_server(exporter_port)
    app_metrics.run_metrics_loop()


if __name__ == "__main__":
    main()
