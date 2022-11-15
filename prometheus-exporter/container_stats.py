import docker
import time


def get_cpu_stats(docker_client, container):
    stats = docker_client.containers.get(container).stats(stream=False)
    UsageDelta = stats['cpu_stats']['cpu_usage']['total_usage'] - stats['precpu_stats']['cpu_usage']['total_usage']
    # from informations : UsageDelta = 25382985593 - 25382168431

    SystemDelta = stats['cpu_stats']['system_cpu_usage'] - stats['precpu_stats']['system_cpu_usage']
    # from informations : SystemDelta = 75406420000000 - 75400410000000

    len_cpu = stats['cpu_stats']['online_cpus']
    # from my informations : len_cpu = 8

    percentage = (UsageDelta / SystemDelta) * len_cpu * 100
    # this is a little big because the result is : 0.02719341098169717

    PUE = 1.4  # Power Useage effectiveness
    power = percentage * len_cpu * PUE / 60

    return power


def get_stats():
    client = docker.DockerClient(base_url='unix:///var/run/docker.sock')
    stats = {}
    for containers in client.containers.list():
        # print(containers.stats(decode=None, stream=False))
        stats[containers.name] = get_cpu_stats(client, containers.id)
    return stats

#
# def run_metrics_loop():
#     """Metrics fetching loop"""
#
#     while True:
#         print(get_stats())
#         time.sleep(5)
#
# run_metrics_loop()
