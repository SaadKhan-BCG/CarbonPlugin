import datetime
from urllib.parse import quote
import requests


# TODO: Error handling for this client
# Common use cases: Http connection failure, location doesnt exist, invalid time
class CarbonEmissions:
    def __init__(self, app_host, app_port, app_url=None):
        self.baseurl = app_url if app_url else f"http://{app_host}:{app_port}"

        # Local dict cache to store carbon values and prevent multiple lookups
        self.region_carbon = {}
        self.region_time_carbon = {}

    def __get_carbon_emissions(self, location, prev_time, to_time):
        """
        Internal method to actually query the Carbon SDK for emissions data

        :param location: Location to extract Carbon Data from
        :param prev_time: url encoded start time
        :param to_time: url encoded end time
        :return: Carbon consumption at given time and region in gCO2eq/watt
        """

        request_url = f"{self.baseurl}/emissions/bylocation?location={location}&time={prev_time}&toTime={to_time}"

        resp = requests.get(url=request_url)
        return resp.json()[0]['rating']

    def get_carbon_emissions_utc(self, location, utc_time):
        """

        :param location: Location to extract Carbon Data from
        :param utc_time: python date object in utc time you want the carbon data at
        :return: Carbon consumption at given time and region in gCO2eq/watt
        """

        hour = utc_time.hour
        carbon = self.region_carbon.get((location, hour))

        if not carbon:
            url_time = quote(utc_time.strftime("%Y-%m-%dT%H:%M"))
            prev_time = utc_time - datetime.timedelta(minutes=1)
            prev_url_time = quote(prev_time.strftime("%Y-%m-%dT%H:%M"))

            carbon = self.__get_carbon_emissions(location, prev_url_time, url_time)
            self.region_carbon[(location, hour)] = carbon
        return carbon

    def get_current_carbon_emissions(self, location):
        carbon = self.region_carbon.get(location)
        if not carbon:
            current_time = datetime.datetime.utcnow()
            carbon = self.get_carbon_emissions_utc(location, current_time)
            self.region_carbon[location] = carbon
        return carbon

    # Empty local cache, used at start of every new refresh
    def clear_carbon_cache(self):
        self.region_carbon = {}
        self.region_time_carbon = {}
