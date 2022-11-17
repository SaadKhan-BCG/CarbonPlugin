import docker
from constants import energy_profile


def get_container_stats(docker_client, container):
    stats = docker_client.containers.get(container).stats(stream=False)

    # CPU Power
    usage_delta = stats['cpu_stats']['cpu_usage']['total_usage'] - stats['precpu_stats']['cpu_usage']['total_usage']
    system_delta = stats['cpu_stats']['system_cpu_usage'] - stats['precpu_stats']['system_cpu_usage']
    len_cpu = stats['cpu_stats']['online_cpus']
    percentage = (usage_delta / system_delta) * len_cpu * 100
    cpu_power = percentage * energy_profile['PUE'] * energy_profile['CPU'] / 3600

    # Memory Power
    mem_usage = stats['memory_stats']['usage'] / 1073741824  # Number is in bytes so divide to get to GB
    mem_power = mem_usage * energy_profile['PUE'] * energy_profile['MEM'] / 3600

    # Network Power
    total_rx = 0
    total_tx = 0
    for _, network in stats['networks'].items():
        total_rx += network['rx_bytes']
        total_tx += network['tx_bytes']

    network_power = (total_tx + total_rx) / 1073741824 * energy_profile['NETWORK'] / 7200
    disk_power = 0  # Usually almost nothing

    # Note total data is in watts/s (CarbonAware SDK works in gCo2/KwH)
    total_power = (cpu_power + mem_power + disk_power + network_power)
    return total_power


def get_stats():
    client = docker.DockerClient(base_url='unix:///var/run/docker.sock')
    stats = {}
    for containers in client.containers.list():
        if not containers.name.startswith("carbon-plugin"):
            stats[containers.name] = get_container_stats(client, containers.id)
    return stats
