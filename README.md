# Transactional Outbox Pattern in Amazon DynamoDB
A demonstration of the transactional outbox messaging pattern (+ Log Trailing) with Amazon DynamoDB (+ Streams) written in Go.

For more information about transaction outbox pattern, please read [this article](https://microservices.io/patterns/data/transactional-outbox.html).

For more information about log trailing pattern, please read [this article](https://microservices.io/patterns/data/transaction-log-tailing.html).

**Requirements**
- 2 tables in Amazon DynamoDB (+ table stream).
- 1 serverless function in Amazon Lambda.
- 1 topic in Amazon Simple Notification Service (SNS).

_Note: Live infrastructure is ready to deploy using the Terraform application from_ `deployments/aws`.

The architecture is very simple as it relies on serverless patterns and services provided by Amazon Web Services.
Most of the heavy lifting is done by Amazon itself. Nevertheless, please consider factors such as 
Amazon Lambda concurrency limits and API calls from/to Amazon services as it could impact both performance and scalability.

Before implementing the solution, please create **two** Amazon DynamoDB tables. One called `students` and 
the other called `outbox`.

The `students` table MUST have a property named `student_id` as _Partition Key_ and another property
named `school_id` as _Sort Key_.

The `outbox` table MUST have a property named `transaction_id` as _Partition Key_ and another property
named `occurred_at` as _Sort Key_.

All keys MUST be **String**.

Finally, enable **Time-To-Live** and **Streams** on the `outbox` table. 

![arch](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/images/StreamsAndTriggers.png)

_Overall architecture, took from [this AWS article](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html)_

The workflow is also very simple and so is straightforward.

1. When writing in the database, the Amazon DynamoDB repository MUST call the `TransactionWriteItems`
   1. Add up business operations into the transaction to any tables required.
   2. Add up the domain events _(encoded using JSON)_ into a single transaction to reduce transaction items and comply 
   with the _25 ops/per-transaction API limit_.
   This item MUST be written in the `outbox table`.
2. The new row will get projected into the `outbox` table stream.
3. The `outbox` table stream will trigger the `log trailing daemon` Amazon Lambda function.
4. The serverless function will:
   1. Decode the domain events from the stream message _(using JSON)_.
   2. Using [`Neutrino Gluon`](https://github.com/NeutrinoCorp/gluon), the `publish()` function will transform
   domain events into integration events (Gluon uses CNCF's CloudEvents specification). 
   3. Gluon will publish the integration events into the desired event bus using the selected codec. In this specific scenario,
   messages will be published in _Amazon Simple Notification Service (SNS)_ using _Apache Avro_ codec.
5. The event bus (Amazon SNS) will route the message to specific destinations. In event-driven systems, the most common
    destination are Amazon Simple Queue Service (SQS) queues (one-queue per job). This is called
    _topic-queue chaining pattern_.

## Additional lecture

### Cleaning the Outbox table

Using the Time-To-Live (TTL) mechanism, batches of messages stored in the `outbox` table will get removed after
a specific time defined by the developer _(for this example, default is 1 day)_. If Amazon DynamoDB is not an option,
a TTL mechanism MUST be implemented manually to keep the `outbox` table lightweight.

An open source alternative is `Apache Cassandra` which based most of its initial implementation from cloud-native solutions such as `Amazon DynamoDB` and `Google BigQuery`. It also has TTL mechanisms out the box.

Furthermore, the defined time for a record to be removed opens the possibility to replay batches of messages generated within transactions.

### Common issues with Event-Driven systems

#### Dealing with duplication and disordered messages when replicating data

As the majority of event-driven systems, messages COULD be published without a specific order
and additionally, messages COULD be published more than one time _(caused by At-least one message delivery)_.

Thus, event handlers operations MUST be idempotent. More in deep:
- When creating an entity/aggregate:
  - If a mutation or deletion arrives first to the system, 
      fail immediately so the process can be retried after a specific time. While this backoff passes by,
      the create operation might get executed.
- When removing an entity/aggregate:
  - If a mutation operation arrives after the removal, the handler MUST return
  no error and acknowledge the arrival of the message.
- When mutating an entity/aggregate:
  - If an old version of the entity/aggregate arrives after, using the
    _Change-Data-Capture (CDC)_ `version` delta or `last_update_time` timestamp, the 
    operation MUST distinct between older versions and skip the actual mutation process and
    acknowledge the message arrival as it was actually updated.
    The `version` field is recommended over `last_update_time` as time precision COULD lead into
    race conditions (e.g. using seconds while the operations take milliseconds, thus, the field could have basically 
    the same time).

If processes keep failing after N-times _(N defined by the developer team)_, store poison messages into a 
Dead-Letter Queue (DLQ) so they can be replayed manually after fixes get deployed. No data should be lost.

#### Dealing with duplication and disordered messages in complex processing

Sometimes, a business might require functionality which align perfectly with the nature of events 
(reacting to change). For example, the product team might require to notify the user when he/she executes a 
specific operation.

In that scenario, using the techniques described before will not be sufficient to deal with the nature of event-driven
systems (duplication and disorder of messages). Nevertheless, they are still solvable as they only require to do a
specific action triggered by some event.

In order to solve duplication of processes, a table named `inbox` (or similar) COULD be used to track message processes 
already executed in the system (even if it is a cluster of nodes).
More in deep:

1. Message arrives.
2. A middleware is called before the actual message process.
   1. The middleware checks if the message was already processed using the message id as key and the table `inbox`.
      1. If message was already processed, stop the process and acknowledge the arrival of the message.
      2. If not, continue with the message processing normally.
3. The message process gets executed.
4. The middleware will be called again.
   1. If the processing was successful, commit the message process as success in the `inbox` table.
5. If processing failed, do not acknowledge the message arrival, so it can be retried.

Finally, one thing to consider while implementing this approach is the necessity of a Time-To-Live (TTL) 
mechanism, just as the `outbox` table, to keep the table lightweight.

Note: This `inbox` table COULD be implemented anywhere as it does not require transactions or any similar mechanism.
It is recommended to use an external database to reduce computational overhead from the main database used by business 
operations. An in-memory database such as Redis _(which also has built-in TTL)_ or even Amazon DynamoDB/Apache Cassandra 
(distributed databases) are one of the best choices as they handle massive read operations efficiently.

In the other hand, if disordered processing is a serious problem for the business, the development team might take advantage
of the previous described approach for duplication of processes adhering workarounds such as the usage of timestamps or 
even deltas to distinct the order of the processes. Getting deeper:

![Correlation and Causation IDs](https://blog-arkency.imgix.net/correlation_id_causation_id_rails_ruby_event/CorrelationAndCausationEventsCommands.png?w=768&h=758&fit=max)

1. Message arrives.
2. A middleware `duplication` is called before the actual message process.
   1. The middleware checks if the process was already processed.
      1. If already processed, stop the process and acknowledge the message.
3. A middleware `disorder` is called before the actual message process.
   1. The middleware verifies if the previous process was already executed using the `causation_id` property.
      1. If not, return an error and do not acknowledge the message, so it can be retried again after a backoff.
      2. If previous process was already executed, continue with the message processing normally.
4. The message process gets executed.
5. The middleware `duplication` will be called again.
    1. If the processing was successful, commit the message process as success in the `inbox` table.
6. If processing failed, do not acknowledge the message arrival, so it can be retried.

For more information about this last approach, please read [this article about correlation and causation IDs](https://blog.arkency.com/correlation-id-and-causation-id-in-evented-systems/).
