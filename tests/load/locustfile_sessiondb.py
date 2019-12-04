from locust import HttpLocust, TaskSet, task
import resource
import random

# https://github.com/locustio/locust/issues/638
resource.setrlimit(resource.RLIMIT_NOFILE, (10240, 9223372036854775807))

features = [
    "hbase_cluster_config",
    "sessiondb_query",
    "enterprise_scale",
    "outlier_filter",
    "countingservice_shard_rangekey_parallelism",
    "age_of_request",
    "test_data_file_update",
    "count_client_feature",
    ]


class UserBehavior(TaskSet):
    headers = {"X-Optimizely-SDK-Key": "SgoHGf3PFyDCJbYksRHiRC"}

    @task(1)
    def listFeatures(self):
        #url = "/users/test-user/features/{}".format(random.choice(features))
        url = "/v1/features/16907921526?outlierFilter=true&earliest=2019-11-06T00%3A34%3A38.872984Z&metrics%5B%5D=UNIQUE&timeout=3000&token=AALvYagCR8IWkHfIx91l2-RIWQhJqOyp&bucket_count=1&layerId=16943330438&resultsAPI=true&goals%5B%5D=16917723650&goals%5B%5D=16907921526&cacheMode=RESULTS_API&groups%5B%5D=COLLECTION&namespace=sessiondb_experiment&end=2019-11-08T15%3A25%3A00.000-08%3A00&visitorEventFirstCounting=true&begin=2019-11-05T16%3A34%3A38.000-08%3A00&featureId=16917423667"
        r = self.client.get(url, headers=self.headers)

class WebsiteUser(HttpLocust):
    task_set = UserBehavior
    min_wait = 5000
    max_wait = 9000
