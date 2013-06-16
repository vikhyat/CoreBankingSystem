CoreBankingSystem
=================

System Architecture
-------------------

We wrote the backend web server in Go, and used Redis as the primary data store. HAProxy was used to load-balance requests to multiple Go web server instances. Redis was used both to store account balances and for locking accounts. Account details were sharded across multiple Redis instances.

The back-end is, therefore, completely scalable since the number of Redis instances and the number of Go web servers can be increased arbitrarily and independently. The only component that cannot be scaled horizontally is the load-balancer.

![Image](../blob/master/arch.png?raw=true)

Performance Evaluation
----------------------

In order to evaluate the performance of the system, we used 5 PCs from the lab, each of which has an i5 processor and 4 GB of RAM. One of these was running a load balancer, and two Redis instances and 3 Go web servers were distributed across the remaining machines.

We ran a script that creates 1000 threads and starts sending random valid requests to the webserver on several machines simultaneously before starting, so that we could identify bottlenecks and scale the system as required.

We started off with a single web server and two Redis instances and gradually started increasing the number of concurrent requests. We found that beyond 800 concurrent requests, the Go webserver was starting to run out of available file descriptors, so we configured HAProxy to not send more than 800 requests to any Go web server. At this time, the Go webserver was only using approximately 8-11% of the CPU so we expect it to be possible to increase the number of concurrent requests the server can handle by increasing the number of file descriptors. However, we were not able to do this because we did not have admin rights to a few of the machines that were running web servers.

Beyond this point, to be able to handle more concurrent requests, we started to add more backend web servers. However after a little while, we encountered another bottleneck -- the network connection between the different machines. Because there were a lot of other users in the lab at that time, we were not able to send more than about 4000 requests at most (~2000 requests on average) per second to the load balancer. As a result, we were not able to fully load-test the system on a large scale. However, because of the fairly low system resource consumption at this point, we conclude that it should be possible for this system to handle at least 10x the number of requests we were able to test.

