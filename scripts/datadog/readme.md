# Docker and datadog agent setup:
[https://github.com/burningion/ecs-fargate-deployment-tutorial](https://github.com/burningion/ecs-fargate-deployment-tutorial)


# Getting started:
[https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_GetStarted.html](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_GetStarted.html)



##Summary of conversation with AWS support:

ECS does not understand docker compose, it only understands Task Definitions,
but there is a way to convert docker compose. One can use the following command to convert docker compose to Fargate compatible Task Definition and then use Task Definition to launch containers/Tasks:
[https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-create.html](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-create.html)

There is an option of launching an ECS service directly from your docker compose file:
[https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-service.html](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-service.html)

Recommended steps:

* Use the ecs-cli compose create command to First Create a Appropriate Task Definition & Go through the fields that are created as part of the Task Definition
* Try to correlate docker compose with Task definition file that got created. That way you will understand what each entry in docker compose got translated to in task definition
* Each Container ( Image ) in Docker file should be converted to an individual Container Definition Section of our Task Defitnition
* Launch a Fargate service from this Task definition

There are 2 different Service Types in ECS:

*  Replica (you want to launch based on availability of nodes)
*  Daemon Set (exactly one instance of a container running on all EC2 instances, useful for logging)

DaemonSet is not applicable for Fargate.
However, Fargate will ensure that Tasks always get placed so, no worries about whether or not container will be launched (as it's AWS Managed)

However, if you are using EC2 Launch Type with ECS
you need to choose Daemon Set to make sure at least one instance of Datadog container runs on all nodes.

