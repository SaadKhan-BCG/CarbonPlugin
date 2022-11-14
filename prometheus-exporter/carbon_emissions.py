import datetime
from urllib.parse import quote
import requests


# TODO: Error handling for this client
# Common use cases: Http connection failure, location doesnt exist, invalid time
class CarbonEmissions:
    def __init__(self, app_host, app_port, app_url=None):
        self.baseurl = app_url if app_url else f"http://{app_host}:{app_port}"

    def __get_carbon_emissions(self, location, prev_time, to_time):
        """

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
        url_time = quote(utc_time.strftime("%Y-%m-%dT%H:%M"))
        prev_time = utc_time - datetime.timedelta(minutes=1)
        prev_url_time = quote(prev_time.strftime("%Y-%m-%dT%H:%M"))

        return self.__get_carbon_emissions(location, prev_url_time, url_time)

    def get_current_carbon_emissions(self, location):
        current_time = datetime.datetime.utcnow()
        return self.get_carbon_emissions_utc(location, current_time)
