# Redis Sentinel Failed Source Label Cleanup Design

## Context

When Redis Sentinel manages failover, Sentinel remains responsible for electing a new source and the Sentinel unit-agent remains responsible for writing that source back to `RedisReplication.spec.source`.

Compose Operator currently rebuilds topology status on every reconciliation. If the node named by `spec.source` cannot be queried or is no longer a Redis source, its topology entry is not healthy. Resource reconciliation nevertheless continues to label that Pod as the read-write node and replaces source host and port labels with empty strings. This can leave the read-write Service selecting the old source and can erase valid bootstrap information from other Redis and Sentinel Pods before Sentinel has written the new source into the spec.

## Goals

- Preserve Sentinel's existing responsibility for updating `spec.source` after failover.
- Stop the read-write Service from selecting an unavailable or demoted source.
- Remove invalid source-routing data from only the old source Pod.
- Preserve ownership information needed to associate the Pod with the `RedisReplication` resource.
- Avoid writing empty source host and port labels to replica and Sentinel Pods.
- Make all label transitions idempotent.

## Non-goals

- Discover or select a new source inside Compose Operator while Sentinel takeover is enabled.
- Change the Sentinel callback or `spec.source` update protocol.
- Change Redis replication commands, Sentinel quorum, or failover timing.
- Delete or restart the failed source Pod.
- Redesign read-write or read-only Services.

## Selected Approach

Use conditional label cleanup for the Pod identified by `spec.source`.

A source is healthy for routing only when its topology entry:

- exists;
- has `status == OK`; and
- has `role == Source`.

When all conditions are satisfied, existing label reconciliation continues and ensures:

- `compose-operator/redis-replication.name=<RedisReplication name>`;
- `compose-operator/redis-replication.readonly=false`;
- `compose-operator/redis-replication.source.host=<announce host>`; and
- `compose-operator/redis-replication.source.port=<announce port>`.

When any condition is not satisfied, Compose Operator removes the following labels from the old source Pod:

- `compose-operator/redis-replication.readonly`;
- `compose-operator/redis-replication.source.host`;
- `compose-operator/redis-replication.source.port`.

It retains `compose-operator/redis-replication.name` so Pod-to-resource mapping and ownership remain available.

The read-write Service selects both the replication name and `readonly=false`. Removing the readonly label therefore removes the failed source from the Service without modifying the Service itself. Until Sentinel writes the elected source into `spec.source`, the Service may intentionally have no ready backend rather than route writes to the wrong Redis node.

## Replica and Sentinel Label Behavior

Replica and Sentinel Pods must not have their source host or port labels overwritten with empty strings merely because the current `spec.source` is unhealthy.

If no healthy source address is available during a reconciliation:

- replica Pods retain their last valid source host and port labels;
- Sentinel Pods retain their last valid source host and port labels; and
- unrelated labels continue to reconcile normally.

After the Sentinel unit-agent updates `spec.source`, the next reconciliation observes the elected source as healthy and writes the new announce host and port to the managed Pods.

## Reconciliation Flow

1. Build topology status from current Redis observations.
2. Evaluate the topology entry associated with `spec.source`.
3. If healthy, reconcile source routing labels normally.
4. If unhealthy, remove the three routing labels from only the old source Pod.
5. Reconcile the read-write Service unchanged; its selector naturally excludes the old source.
6. Do not replace replica or Sentinel source labels with empty values.
7. Wait for Sentinel to elect a source and update `spec.source` through the existing unit-agent callback.
8. On a later reconciliation, label the new source and propagate its announce address.

## Error Handling

- A missing old source Pod is reported through the existing resource reconciliation condition and event path; no replacement Pod is created by Compose Operator.
- Kubernetes update conflicts follow the controller-runtime retry/requeue behavior already used by the controller.
- Removing an already absent label is a no-op and must not trigger a Pod update.
- Failure to update one Pod must not cause valid labels on other Pods to be cleared.

## Tests

Add focused unit tests covering:

1. A healthy source receives the normal read-write and source address labels.
2. An unreachable source loses exactly the readonly, source host, and source port labels.
3. The replication name label remains on an unhealthy source.
4. Repeating cleanup on an already-clean Pod is idempotent.
5. A node whose observed role is Replica is cleaned even if it is still named by `spec.source`.
6. After `spec.source` changes to the elected source, the new source receives `readonly=false` and its announce address.
7. Replica source labels retain their previous valid values while `spec.source` is unhealthy.
8. Sentinel source labels retain their previous valid values while `spec.source` is unhealthy.
9. The read-write Service selector no longer matches the cleaned old source Pod.

Controller integration coverage should simulate the sequence source healthy, source unavailable, Sentinel spec update, and new source healthy. The test must assert that Compose Operator never chooses a new source itself.

## Operational Result

During the failover interval, writes may temporarily have no Kubernetes Service endpoint. This is intentional fail-closed behavior. Once Sentinel has updated `spec.source`, Compose Operator restores the read-write endpoint to the elected source without changing the existing Sentinel ownership model.
