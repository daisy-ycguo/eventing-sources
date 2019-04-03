# RabbitMQ - Source

The RabbitMQ Event source enables Knative Eventing integration with RabbitMQ. When a message is published to RabbitMQ exchanges, the RabbitMQ Event Source will consume the message and post that message the corresponding event sink.

This sample demonstrates how to configure, deploy, and use the RabbitMQ Event Source with a Knative Service.

## Build and Deploy Steps

### Prerequisites

1. An existing instance of RabbitMQ must be running to use the RabbitMQ
   Event Source.
   - In order to publish and produce messages, a exchange must be created on the
     RabbitMQ.
   - A broker corresponding to RabbitMQ instance, which support amqp protocal, must be obtained.
2. Install the `ko` CLI for building and deploying purposes.
   ```
   go get github.com/google/go-containerregistry/cmd/ko
   ```
3. A container registry, such as a Docker Hub account, is required.
   - Export the `KO_DOCKER_REPO` environment variable with a value denoting the
     container registry to use.
     ```
     export KO_DOCKER_REPO="docker.io/YOUR_USER/YOUR_REPO"
     ```

### Build and Deployment

The following steps build and deploy the RabbitMQ Event Controller, Source,
and an Event Display Service.

- Assuming current working directory is the project root `eventing-sources`.

#### RabbitMQ Event Controller

1. Build the RabbitMQ Event Controller and configure a Service Account,
   Cluster Role, Controller, and Source.
   ```
   $ ko apply -f contrib/rabbitmq/config
   ...
   serviceaccount/rabbitmq-controller-manager created
   clusterrole.rbac.authorization.k8s.io/eventing-sources-rabbitmq-controller created
   clusterrolebinding.rbac.authorization.k8s.io/eventing-sources-rabbitmq-controller created
   customresourcedefinition.apiextensions.k8s.io/rabbitmqsources.sources.eventing.knative.dev configured
   service/rabbitmq-controller created
   statefulset.apps/rabbitmq-controller-manager created
   ```
2. Check that the `rabbitmq-controller-manager-0` pod is running.
   ```
   $ kubectl get pods -n knative-sources
   NAME                         READY     STATUS    RESTARTS   AGE
   rabbitmq-controller-manager-0   1/1       Running   0          42m
   ```
3. Check the `rabbitmq-controller-manager-0` pod logs.
   ```
   $ kubectl logs rabbitmq-controller-manager-0 -n knative-sources
   2019/04/03 09:47:33 Registering Components.
   2019/04/03 09:47:33 Setting up Controller.
   2019/04/03 09:47:33 Adding the RabbitMQ Source controller.
   2019/04/03 09:47:33 Starting RabbitMQ controller.
   ```

#### Event Display

1. Build and deploy the Event Display Service.
   ```
   $ ko apply -f contrib/rabbitmq/samples/event-display.yaml
   ...
   service.serving.knative.dev/event-display created
   ```
2. Ensure that the Service pod is running. The pod name will be prefixed with
   `event-display`.
   ```
   $ kubectl get pods
   NAME                                            READY     STATUS    RESTARTS   AGE
   event-display-00001-deployment-5d5df6c7-gv2j4   2/2       Running   0          72s
   ...
   ```

#### RabbitMQ Event Source

1. Modify `contrib/rabbitmq/samples/event-source.yaml` accordingly with bootstrap
   servers, topics, etc...
  - You can also use the RabbitMQ statefulset in `contrib/rabbitmq/samples/rabbitmq-statefulset.yaml`.
2. Build and deploy the event source.
   ```
   $ kubectl apply -f contrib/rabbitmq/samples/event-source.yaml
   ...
   rabbitmqsource.sources.eventing.knative.dev/rabbitmq-source created
   ```
3. Check that the event source pod is running. The pod name will be prefixed
   with `rabbitmq-source`.
   ```
   $ kubectl get pods
   NAME                                  READY     STATUS    RESTARTS   AGE
   rabbitmq-source-xlnhq-5544766765-dnl5s   1/1       Running   0          40m
   ```
4. Ensure the RabbitMQ Event Source started with the necessary
   configuration.
   ```
   $ kubectl logs rabbitmq-source-xlnhq-5544766765-dnl5s
   {"level":"info","ts":"2019-04-03T19:23:31.998Z","caller":"receive_adapter/main.go:92","msg":"Starting RabbitMQ Receive Adapter...","adapter":{"AMQPBroker":"","ExchangeName":"","SinkURI":"http://event-display.default.svc.cluster.local/","Net":{"SASL":{"Enable":true,"User":"guest","Password":"guest"},"TLS":{"Enable":false}}}}
   {"level":"info","ts":1554319411.9983,"logger":"fallback","caller":"adapter/adapter.go:91","msg":"Starting with config: {adapter 22 0  0xc0001379d0}"}
   {"level":"info","ts":1554319411.9983342,"logger":"fallback","caller":"adapter/adapter.go:101","msg":"Connecting to RabbitMQ on : amqp://guest:guest@rabbitmq/"}
   ```

### Verify

1. Produce the message shown below to RabbitMQ.
   ```
   {"msg": "This is a test!"}
   ```
2. Check that the RabbitMQ Event Source consumed the message and sent it to
   its sink properly.
   ```
   $ kubectl logs rabbitmq-source-xlnhq-5544766765-dnl5s
   ...
   {"level":"info","ts":1553034726.5351956,"logger":"fallback","caller":"adapter/adapter.go:121","msg":"Received: {value 15 0 {\"msg\": \"This is a test!\"} <nil>}"}
   {"level":"info","ts":1553034726.546107,"logger":"fallback","caller":"adapter/adapter.go:154","msg":"Successfully sent event to sink"}
   ```
3. Ensure the Event Display received the message sent to it by the Event Source.

   ```
   $ kubectl logs event-display-00001-deployment-5d5df6c7-gv2j4 -c user-container

   ☁️  CloudEvent: valid ✅
   Context Attributes,
     SpecVersion: 0.2
     Type: dev.knative.rabbitmq.event
     Source: excha
     ID: 00001121322
     Time: 2019-03-19T22:32:06.535321588Z
     ContentType: application/json
     Extensions:
       key:
   Transport Context,
     URI: /
     Host: event-display.default.svc.cluster.local
     Method: POST
   Data,
     {
       "msg": "This is a test!"
     }
   ```

## Teardown Steps

1. Remove the RabbitMQ Event Source
   ```
   $ ko delete -f contrib/rabbitmq/samples/source.yaml
   rabbitmqsource.sources.eventing.knative.dev "rabbitmq-source" deleted
   ```
2. Remove the Event Display
   ```
   $ ko delete -f contrib/rabbitmq/samples/event-display.yaml
   service.serving.knative.dev "event-display" deleted
   ```
3. Remove the RabbitMQ Event Controller
   ```
   $ ko delete -f contrib/rabbitmq/config
   serviceaccount "rabbitmq-controller-manager" deleted
   clusterrole.rbac.authorization.k8s.io "eventing-sources-rabbitmq-controller" deleted
   clusterrolebinding.rbac.authorization.k8s.io "eventing-sources-rabbitmq-controller" deleted
   customresourcedefinition.apiextensions.k8s.io "rabbitmqsources.sources.eventing.knative.dev" deleted
   service "rabbitmq-controller" deleted
   statefulset.apps "rabbitmq-controller-manager" deleted
   ```
