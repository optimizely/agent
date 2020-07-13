---
title: "Optimizely Agent"
excerpt: ""
slug: "optimizely-agent"
hidden: false
metadata: 
  title: "Optimizely Agent microservice - Optimizely Full Stack"
createdAt: "2020-02-21T20:35:58.387Z"
updatedAt: "2020-04-01T20:51:52.458Z"
---
Optimizely Agent is a stand-alone, open-source, and highly available microservice that provides major benefits over using Optimizely SDKs in certain use cases. The Agent [REST API](https://library.optimizely.com/docs/api/agent/v1/index.html) offers consolidated and simplified endpoints for accessing all the functionality of Optimizely Full Stack SDKs. 

A typical production installation of Optimizely Agent is to run two or more services behind a load balancer or proxy. The service itself can be run via a Docker container or installed from source. See [Setup Optimizely Agent](doc:setup-optimizely-agent) for instructions on how to run Optimizely Agent.

### Example Implementation
![example implementation](https://raw.githubusercontent.com/optimizely/agent/master/docs/images/agent-example-implementation.png)
# Should I Use Optimizely Agent?

Here are some of the top reasons to consider using Optimizely Agent:

## 1. Service Oriented Architecture (SOA)
If you already separate some of your logic into services that might need to access the Optimizely decision APIs, we recommend using Optimizely Agent. 

The images below compare implementation styles in a service-oriented architecture, first *without* using Optimizely Agent, which shows six SDK embedded instances:

!["A diagram showing the use of SDKs installed on each service in a service oriented architecture \n(Click to Enlarge)"](https://raw.githubusercontent.com/optimizely/agent/master/docs/images/agent-service-oriented-architecture.png)

Now *with* Agent, instead of installing the SDK six times, you create just one Optimizely instance: an HTTP API that every service can access as needed. 

!["A diagram showing the use of Optimizely Agent in a single service \n(Click to Enlarge)"](https://raw.githubusercontent.com/optimizely/agent/master/docs/images/agent-single-service.png)

## 2. Standardize Access Across Teams
If you want to deploy Optimizely Full Stack once, then roll out the single implementation across a large number of teams, we recommend using Optimizely Agent. 

By standardizing your teams' access to the Optimizely service, you can better enforce processes and implement governance around feature management and experimentation as a practice.

!["A diagram showing the central and standardized access to the Optimizely Agent service across an arbitrary number of teams.\n(Click to Enlarge)"](https://raw.githubusercontent.com/optimizely/agent/master/docs/images/agent-standardized-access.png)

## 3. Networking Centralization
You don’t want many SDK instances connecting to Optimizely's cloud service from every node in your application. Optimizely Agent centralizes your network connection. Only one cluster of agent instances connects to Optimizely for tasks  like update [datafiles](doc:get-the-datafile) and dispatch [events](doc:track-events).

## 4. Languages
You’re using a language that isn’t supported by a native SDK (i.e. Elixir, Scala, Perl). While its possible to create your own service using an Optimizely SDK of your choice, you could also customize the open-source Optimizely Agent to your needs without building the service layer on your own. 


# Reasons to *not* use Optimizely Agent
If your use case wouldn't benefit greatly from Optimizely Agent, you should consider the below reasons to *not* use Optimizely Agent and review Optimizely's many [open-source SDKs](doc:sdk-reference-guides) instead. 

## 1. Latency
If time to provide bucketing decisions is a primary concern for you, you may want to use an embedded Full Stack SDK rather than Optimizely Agent. 
| Implementation Option | Decision Latency |
|-----------------------|------------------|
| Embedded SDK          | microseconds     |
| Optimizely Agent      | milliseconds     |
## 2. Monolith
If your app is constructed as a monolith, embedded SDKs might be easier to install and might be a more natural fit for your application and development practices. 

## 3. Velocity
If you’re looking for the fastest way to get a single team up and running with deploying feature management and experimentation, embedding an SDK is the best option for you at first. You can always start using Optimizely Agent later, and it can even be used alongside Optimizely Full Stack SDKs running in another part of your stack.

# Best Practices
While every implementation is different, you can review this section of best practices for tips on these commonly discussed topics.


## How many Agent instances should I deploy?
Agent can scale to large decision / event tracking volumes with relatively low CPU / memory specs. For example, at Optimizely, we scaled our deployment to 740 clients with a cluster of 12 agent instances, which in total use 6 vCPUs and 12GB RAM. You will likely need to focus more on network bandwidth than compute power.

## Using a load balancer
Any standard load balancer should let you route traffic across your agent cluster. At Optimizely, we used an AWS Elastic Load Balancer (ELB) for our internal deployment. This allows us to transparently scale our agent cluster as internal demands increase.

## Synchronizing datafiles across Agent instances
Agent offers eventual rather than strong consistency across datafiles.
In detail, today, each agent instance maintains a dedicated, separate cache. Each agent instance persists an SDK instance for each SDK key your team uses.  Agent instances automatically keep datafiles up to date for each SDK key instance so that you will have eventual consistency across the cluster. The rate of the datafile update can be [set as the configuration](doc:configure-optimizely-agent) value ```OPTIMIZELY_CLIENT_POLLINGINTERVAL```  (the default is 1 minute).
Because SDKs are generally stateless today, they shouldn’t need to share data. We plan to add a common backing data store, so we invite you to share your feedback. 
If you require strong consistency across datafiles, then we recommend an active / passive deployment where all requests are made to a single vertically scaled host, with a passive, standby cluster available for high availability.
