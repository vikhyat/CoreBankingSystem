CoreBankingSystem
=================

Introduction
------------

Core Banking is a collection of services provided by a group of networked bank branches. Bank customers may access their funds and other simple transactions from any of the member branch offices. CORE stands for Centralized Online Realtime Environment. Core Banking is normally defined as the business conducted by a banking institution with its retail and small business customers. Core Banking basically is depositing and lending of money. Nowadays, most banks use Core Banking applications to support their operations.

This basically means that all of the bank's branches access applications from centralized datacenters. This also means that the deposits made are reflected immediately on the bank's servers and the customer can withdraw the deposited money from any of the bank's branches throughout the world. A few decades ago it used to take at least a day for a transaction to reflect in the account because each branch had their local servers, and the data from the server in each branch was sent in a batch to the servers in the datacenter only at the end of the day.

The application simulates a core banking system that attempts to support banking services for a large number of customers (about 10 million). Normal core banking functions will include deposit accounts, loans, mortgages and payments. Banks make these services available across multiple channels like ATMs, Internet banking, and branches. In this application, only the following 3 banking operations are implemented -– **Withdraw**, **Deposit** and **Transfer**.

The system is built with concurrency and multicore systems in mind. It uses an optimistic locking protocol to achieve thread safety in critical sections.

Software Stack
--------------

### Redis (Data Store)

Redis is an open-source, networked, in-memory, key-value data store with optional durability. The Redis data model is a dictionary where keys are mapped to values. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets. Redis supports high level, atomic, server side set operations and sorting.

Redis typically holds the whole dataset in memory. Persistence is reached by snapshotting, where the dataset is asynchronously transferred from memory to disk from time to time.

Redis supports master-slave replication. Data from any Redis server can replicate to any number of slaves. A slave may be a master to another slave. This allows Redis to implement a single-rooted replication tree. Redis slaves are writable, permitting intentional and unintentional inconsistency between instances.

When the durability of data is not needed, the in-memory nature of Redis allows it to perform extremely well compared to database systems that write every change to disk before considering a transaction committed. There is no notable speed difference between write and read operations.

### Go (Web Server)

Go is an open source programming environment that makes it easy to build simple, reliable, and efficient software. Go is expressive, concise, clean, and efficient. Its concurrency mechanisms make it easy to write programs that get the most out of multicore and networked machines, while its novel type system enables flexible and modular program construction.

Go compiles quickly to machine code yet has the convenience of garbage collection and the power of run-time reflection. It's a fast, statically typed, compiled language that feels like a dynamically typed, interpreted language.

### HAProxy (Load Balancer)

HAProxy is a free, very fast and reliable solution offering high availability, load balancing, and proxying for TCP and HTTP-based applications. It is particularly suited for web sites crawling under very high loads while needing persistence or application layer processing, to improve the performance by spreading requests across multiple servers. It has a reputation for being fast, efficient (in terms of processor and memory usage) and stable.

HAProxy implements an event-driven, single-process model which enables support for very high number of simultaneous connections at very high speeds. Hence, for best results on multi-processor systems, the transaction handlers must be optimized to get the most work done from every CPU cycle.


System Architecture
-------------------

We wrote the backend web server in Go, and used Redis as the primary data store. HAProxy was used to load-balance requests to multiple Go web server instances. Redis was used both to store account balances and for locking accounts. Account details were sharded across multiple Redis instances.

The back-end is, therefore, completely scalable since the number of Redis instances and the number of Go web servers can be increased arbitrarily and independently. The only component that cannot be scaled horizontally is the load-balancer.

![Image](../master/doc/arch.png?raw=true)

Application Layer Architecture
------------------------------

The banking system implements the following three operations on accounts in the Go webserver frontend:

#### Withdraw

In: **{account, amount}** Out: **{balance}**

This operation involves the decrease in account balance for the **account**, specified in the HTTP request. It acquires a lock on the **account**, and defers its release until the operation completes successfully. In the critical section, the operation reduces the account balance by **amount** using the Redis operation **HINCRBY**.

#### Deposit

In: **{account, amount}** Out: **{balance}**

This operation involves the increase in account balance for the **account**, specified in the HTTP request. It acquires a lock on the **account**, and defers its release until the operation completes successfully. In the critical section, the operation raises the account balance by **amount** using the Redis operation **HINCRBY**.

#### Transfer

In: **{source, destination, amount}** Out: **None**

This operation involves the decrease in account balance for the **source** account, and a simultaneous increase for the **destination** account, specified in the HTTP request. It acquires locks on both the **source** and the **destination**, and defers their release until the operation completes successfully. In the critical section, the operation decreases the **source** balance by **amount**, and increases the **destination** balance by **amount** using the Redis operation **HINCRBY**.

Performance Evaluation
----------------------

In order to evaluate the performance of the system, we used 5 PCs from the lab, each of which has an i5 processor and 4 GB of RAM. One of these was running a load balancer, and two Redis instances and 3 Go web servers were distributed across the remaining machines.

We ran a script that creates 1000 threads and starts sending random valid requests to the webserver on several machines simultaneously before starting, so that we could identify bottlenecks and scale the system as required.

We started off with a single web server and two Redis instances and gradually started increasing the number of concurrent requests. We found that beyond 800 concurrent requests, the Go webserver was starting to run out of available file descriptors, so we configured HAProxy to not send more than 800 requests to any Go web server. At this time, the Go webserver was only using approximately 8-11% of the CPU so we expect it to be possible to increase the number of concurrent requests the server can handle by increasing the number of file descriptors. However, we were not able to do this because we did not have admin rights to a few of the machines that were running web servers.

Beyond this point, to be able to handle more concurrent requests, we started to add more backend web servers. However after a little while, we encountered another bottleneck -- the network connection between the different machines. Because there were a lot of other users in the lab at that time, we were not able to send more than about 4000 requests at most (~2000 requests on average) per second to the load balancer. As a result, we were not able to fully load-test the system on a large scale. However, because of the fairly low system resource consumption at this point, we conclude that it should be possible for this system to handle at least 10x the number of requests we were able to test.

References
----------

1. Parallel Programming for Multicore and Cluster Systems - Rauber & Runger
2. [Core Banking](http://en.wikipedia.org/wiki/Core_banking)
3. [HAProxy - The Reliable, High Performance TCP/HTTP Load Balancer](http://haproxy.1wt.eu)
4. [Redis](http://redis.io)
5. [The Go Programming Language](http://golang.org/)
6. [Introducing the Go net/http package (an interlude)](http://golang.org/doc/articles/wiki/)
7. [Redigo, a Go client for the Redis database](https://github.com/garyburd/redigo)

