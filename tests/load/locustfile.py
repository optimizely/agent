from locust import HttpLocust, TaskSet, task
import resource
import random

# https://github.com/locustio/locust/issues/638
resource.setrlimit(resource.RLIMIT_NOFILE, (10240, 9223372036854775807))

features = [
    "hbase_cluster_config_feature",
    "hbase_cluster_config",
    "sessiondb_query",
    "enterprise_scale",
    "outlier_filter",
    "countingservice_shard_rangekey_parallelism",
    "hbase_cluster_config_canary",
    "age_of_request",
    "test_data_file_update",
    "count_client_feature",
    ]


class UserBehavior(TaskSet):
    headers = {"X-Optimizely-SDK-Key": "SgoHGf3PFyDCJbYksRHiRC"}

    @task(1)
    def listFeatures(self):
        url = "/users/test-user/features/{}".format(random.choice(features))
        self.client.get(url, headers=self.headers)

class WebsiteUser(HttpLocust):
    task_set = UserBehavior
    min_wait = 5000
    max_wait = 9000
