# Peanut Kubernetes Pipelines

Peanut is an operator for running pipelines in Kubernetes.

## What is a pipeline?

In its current form, a PeanutPipeline is a set of stages that take the form of a Directed Acyclic Graph (https://en.wikipedia.org/wiki/Directed_acyclic_graph)

This means that stages in a pipeline can depend on other stages and will only be executed if parent stages were successful.   

## What is a stage?

A stage is a template for a Kubernetes Job that the operator will end up scheduling on the cluster.

## Deploying the operator

TODO

## Examples

Creating a PeanutStage:
```yaml
apiVersion: pipeline.peanut.dev/v1alpha1
kind: PeanutStage
metadata:
  name: stage-sample
spec:
  image: busybox
  command: ["sleep"]
  args: ["10"]
---
apiVersion: pipeline.peanut.dev/v1alpha1
kind: PeanutStage
metadata:
  name: stage-sample-two
spec:
  image: busybox
  command: ["sleep"]
  args: ["15"]
```  

Creating a PeanutPipeline:
```yaml
apiVersion: pipeline.peanut.dev/v1alpha1
kind: PeanutPipeline
metadata:
  name: pipeline-sample
spec:
  stages:
  - name: "First Stage"
    type: stage-sample
  - name: "Second Stage"
    type: stage-sample-two
  - name: "Third Stage"
    type: stage-sample
    dependsOn:
    - "Second Stage"
  - name: "Fourth Stage"
    type: stage-sample
    dependsOn:
      - "Second Stage"
  - name: "Fifth Stage"
    type: stage-sample-two
    dependsOn:
      - "Third Stage"
      - "Fourth Stage"
```  

This creates the following pipeline
```
           +---------+
           |         |
           |  First  |
       +-->+  Stage  |
       |   |         |
       |   +---------+
+------+
       |   +--------+      +---------+
       |   |        |      |         |
       +-->+ Second |      |  Third  +----+
           | Stage  +--+-->+  Stage  |    |
           |        |  |   |         |    |   +---------+
           +--------+  |   +---------+    |   |         |
                       |                  +-->+  Fifth  |
                       |   +--------+     |   |  Stage  |
                       |   |        |     |   |         |
                       |   | Fourth |     |   +---------+
                       +-->+ Stage  +-----+
                           |        |
                           +--------+
```

Executing our pipeline:
```yaml
apiVersion: pipeline.peanut.dev/v1alpha1
kind: PeanutPipelineExecution
metadata:
  name: peanutpipelineexecution-sample
spec:
  pipelineName: pipeline-sample
```

The operator will schedule the jobs according to the definition created.

If a stage fails all stages in the same branch won't be scheduled. 

At the moment there isn't much flexibility on the behaviour of the pipeline when a stage fails. 

## But Why?

![](https://media.giphy.com/media/s239QJIh56sRW/giphy.gif)

I started this project to learn how to develop an operator for k8s using kubebuilder.

## Can I use this in production?

I would strongly advise against it, use at your own risk. This is very much a work in progress and not well tested.

## Contributing

Feel free to fork the project and submit a PR

