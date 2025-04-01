
# Metadata Broker

We need two different kind of data sharing:

1. the User Interface (whether is graphical or API based) needs to expose aggregate information to the user
   1. eg: when looking at a Host I want to know which Apps are running on it
2. each component needs to know what are the labels available in the system
   1. eg: when creating a CA Deployment the user want's to know the existing labels

At the same time we have requirements to keep the 3 subsystems independent so that they can be used in isolation (this is a future feature, but we need to make sure we don't cross boundaries, or it will be very hard to disentangle the components later on).

To address the above requirements, we've created the metadata-broker service to store and retrive metadata entered by the user, accross services, enabling the orch-ui to populate the UI to recommend metadata keys or values to the user, based on previous entries.

The Metadata broker must keep metadata as isolated as possible from the client components, the Broker **should only store the metadata key names and values that can be used in the system, not their association with objects in the data model**. Storing such information should be up to the client component. The Broker exposes a CRUD API interface that can be used to control the data lifecycle, this should be made available through an Administrative Dashboard and controlled via RBAC. For example, if a metadata with a typo is stored the administrator can remove it via the dashboard, and when the metadata is corrected in the subcomponent the new (and correct) value will be stored in the broker. **It is important to note that metadata creation should be idempotent**, creating the same metadata multiple times will not change the state in the broker.

The Metadata broker is isolated per project.  Users can only see metadata for the project they are in.

### High Level Architecture

```mermaid
    C4Container
    title Container diagram for Internet Banking System

    Container_Boundary(c1, "Metadata Broker") {
        Container(mb, "API Server", "Golang", "Provides the API interface to interact with system-wide Metadata")
        ContainerDb(db, "Database", "JSON Object Store", "Stores the metadata")
        Rel(mb, db, "", "")
    }

    Container_Ext(lp-o, "App Orch", "", "")
    Rel(lp-o, mb, "", "")
    Container_Ext(lp-c, "Cluster Orch", "", "")
    Rel(lp-c, mb, "", "")
    Container_Ext(infra, "Edge Infra", "", "")
    Rel(infra, mb, "", "")
```

Each component will know how to reach the Broker in order to keep it up to date with the existing metadata.
The high-level call flow when something is created/updated in a component is as follows:

1. The (pertinent*) metadata are sent to the Broker
2. The object is saved in the component store, along with its own metadata

_* There are cases in which not all the metadata stored with an object need to be shared outside the sub-system boundaries, in that case is perfectly fine for the component not to share them with the Broker._

### End to End usage examples

#### Package deployment

Following are demonstrated some example flows, involving the User/Operator and the high-level components. This document does not drill down into how the sub-systems would implement the calls to the broker as that’s an internal detail. The following example loosely follow the ZTP flow planned for WG as demonstrated in Milestone, and assumes to start from an empty system.

> NOTE: in the following set of examples `customer` is used as metadata, but it's a symbolic value that can replaced with anything that has value for the operator. The same exact pattern applies if multiple metadata are used to identify elements.

##### Creation of a Package Deployment

```mermaid
sequenceDiagram
    autonumber
    participant client as Client
    participant infra as Edge Infra
    participant lpo as App Orch
    participant mdb as Metadata Broker

    mdb->>client: GetSystemMetadata
    Note over mdb,client: Receives all the available metadata
    client->>lpo: Create CaDeployment(customer=culvers)

    lpo->>mdb: CreateOrUpdateMetadata(customer=culvers)
    lpo->>lpo: Stores the CaDeployment and related labels
```

1. The client fetches all the existing Metadata from the broker
    1. Nothing is returned since nothing associated to a label exists in the system yet
2. The User creates a Deployment with a label of “customer=culvers”
3. App Orch sends the metadata to the Broker
4. Once the metadata are stored in the broker, AppOrch stores the data internally and proceeds
    1. This can be seen as a "validation" step for metadata, which is planned feature in the Broker

##### Allocation of a Host to a specific Site

_We assume that Region and a Site are already present in the system. Host provisioning is not depicted as out of scope for this document._

```mermaid
sequenceDiagram
    autonumber
    participant client as Client
    participant infra as Edge Infra
    participant lpo as App Orch
    participant mdb as Metadata Broker

    mdb->>client: GetSystemMetadata
    Note over mdb,client: Receives all the available metadata
    client->>infra: Allocate a host to a Site (customer=culvers, site=site-xyz, region=region-abc)
    infra->>mdb: CreateOrUpdateMetadata(region=Us-West, site=Portland, customer=culvers)
    Note over infra,mdb: Since everything already exist, nothing is added to the broker
```

1. The client fetches all the existing Metadata from the broker
    1. The “customer” metadata is returned by the broker along with all the existing values
2. The User allocates the Host to the Site (see workflow in Demo1)
    1. The “customer” label is proposed to the used, with a selection of the existing values
    2. New values can be added to the “customer” metadata
    3. Other custom metadata can be added to the Host
3. Edge Infra internally converts the Host data to a set of metadata that wants to store in the broker so that other components can benefit from it
    1. Note that the internal value of “site=site-xyz” is rendered to a human-readable metadata as “site=Portland”
At this point if we go back to create a second deployment we’ll be suggested three different labels that we can use: Region, Site and Customer.

#### Cluster Management

This workflow is used to demonstrate that the proposed implementation of the Metadata Broker provides flexibility in support of other use-cases.

##### Creation of a Cluster Template

```mermaid
sequenceDiagram
    autonumber
    participant client as Client
    participant infra as Edge Infra
    participant lpc as Cluster Orch
    participant mdb as Metadata Broker

    mdb->>client: GetSystemMetadata
    Note over mdb,client: Receives all the availabel metadata

    client->>lpc: Create a Cluster Template (name=ran-node, ...)
    lpc->>mdb: CreateOrUpdateMetadata(clusterTemplate=ran-node)
```

When a Cluster Template is created in Cluster Orch (how it is created is outside the scope of this document), a clusterTemplate metadata is registered in the Broker

##### Creation of a new k8s cluster

```mermaid
sequenceDiagram
    autonumber
    participant client as Client
    participant infra as Edge Infra
    participant lpc as Cluster Orch
    participant mdb as Metadata Broker

    mdb->>client: GetSystemMetadata
    Note over mdb,client: Receives all the available metadata
    client->>infra: Allocate a host to a Site (clusterTemplate=ran-node, ...)
    infra->>mdb: CreateOrUpdateMetadata(clusterTemplate=ran-node, ...)
    Note over infra,mdb: Since everything already exist, nothing is added to the broker

    infra->>lpc: Trigger the cluster command
    lpc->>mdb: CreateOrUpdateMetadata(clusterName=portland-cluster1, ...)
```

> NOTE The cluster creation is actually triggered from Node Agent to Cluster Agent, that details has been omitted as not relevant for this diagram

1. The client fetches all the existing Metadata from the broker
    1. The clusterTemplate=ran-node is reported
2. The Host is allocated to a Site
    1. The clusterTemplate=ran-node label is suggested to the user, and assigned to the Host
3. The metadata are sent to the Broker
4. The cluster creation is triggered (this happens via the agents on the Edge Node)
    1. No clusterName label was set, so a name is autogenerated and a new cluster is created
5. The clusterName label is stored in the broker

##### Adding a Host (node) to an existing cluster

```mermaid
sequenceDiagram
    autonumber
    participant client as Client
    participant infra as Edge Infra
    participant lpc as Cluster Orch
    participant mdb as Metadata Broker

    mdb->>client: GetSystemMetadata
    Note over mdb,client: Receives all the available metadata
    client->>infra: Allocate a host to a Site (clusterTemplate=ran-node, clusterName=portland-cluster1)
    infra->>mdb: CreateOrUpdateMetadata(clusterTemplate=ran-node, clusterName=portland-cluster1)
    Note over infra,mdb: Since everything already exist, nothing is added to the broker

    infra->>lpc: Trigger the cluster command*
    lpc->>mdb: CreateOrUpdateMetadata(clusterTemplate=ran-node, clusterName=portland-cluster1)
```

This flow is extremely similar to the previous one, with the difference that when at step 1 the client fetches all the available metadata, the clusterName=portland-cluster1 is returned and suggested to the User. When this is added to the Host, and thus reported to Cluster Orch, the server is added to the already existing cluster.

### Future Features

It would be easy, and already valuable, to implement a “blind” storage of metadata. But there is room for future expansions to provide even more value:

- Provide RBAC to the metadata (eg: only the Admins can create new labels in the system)
- Provide validation (eg: if label testbed is added, cannot add label environment=production)
- Provide an administrative UI to pre-populate labels in the system (this could be embedded in the LP UI)
- Support for custom(er) defined hierarchies'

## Metadata Aggregator

The metadata aggregator would not store any data but act as an API gateway so that when the client request the existing metadata, it will query each sub-system and return the aggregated data.

The above provided examples would still be supported, following is a diagram adapting one single example to describe the call flow:

```mermaid
sequenceDiagram
    participant client as Client
    participant infra as Edge Infra
    participant lpc as Cluster Orch
    participant lpo as App Orch
    participant mda as Metadata Aggregator

    client->>mda: GetSystemMetadata
    mda->>infra: List Hosts
    mda->>lpc: List Clusters
    mda->>lpo: List Deployments
    mda->>client: return GetSystemMetadata
```
