<p align="center">
  <a href="https://github.com/tkestack/knitnet-operator">
    <img src="https://github.com/tkestack/knitnet-operator/workflows/CI%20Pipeline/badge.svg" alt="Github CI">
  </a>
  <a href="https://goreportcard.com/report/github.com/tkestack/knitnet-operator">
    <img src="https://goreportcard.com/badge/github.com/tkestack/knitnet-operator" alt="GoReportCard">
  </a>
  <a href="https://quay.io/repository/danielxlee/knitnet-operator">
    <img src="https://img.shields.io/badge/container-ready-green" alt="Docker">
  </a>
  <a href="https://github.com/tkestack/knitnet-operator/master/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License">
  </a>
</p>

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Knitnet Operator](#knitnet-operator)
  - [Architecture](#architecture)
    - [Purpose](#purpose)
    - [Supported Features](#supported-features)
  - [Getting Started](#getting-started)
    - [Example](#example)
    - [Prerequisites](#prerequisites)
    - [Quickstart](#quickstart)
    - [Verify](#verify)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Knitnet Operator

A Golang based knitnet operator that will make/oversee Submariner components on top of the Kubernetes.

## Architecture

<div align="center">
    <img src="./docs/icons/submariner-arch.png">
</div>

### Purpose

The purpose of creating this operator was to provide an easy and production-grade setup of Submariner components on Kubernetes. It doesn't care if you have a plain on-prem Kubernetes or cloud-based.

### Supported Features

Here the features which are supported by this operator:-

- Deploy submariner broker
- Join cluster to broker
- Check k8s server version
- Support cloud prepare (aws, gcp)
- Support lighthouse, globalnet enable/disable

## Getting Started

### Example

The configuration of Knitnet setup should be described in Knitnet CRD. You will find all the examples manifests in [example](./config/samples) folder.

### Prerequisites

Knitnet operator requires a Kubernetes cluster of version `>=1.15.0`. If you have just started with Operators, its highly recommended to use latest version of Kubernetes. And the prepare 2 cluster, example `cluster-a` and `cluster-b`

### Quickstart

The setup can be done by using `kustomize`.

1. Clone source code

    ```shell
    git clone https://github.com/tkestack/knitnet-operator.git
    cd knitnet-operator
    ```

1. Deploy broker

    - Install knitnet operator

        Switch to `cluster-a`

        ```shell
        kubectl config use-context cluster-a
        ```

        Deploy operator

        ```shell
        make deploy
        ```

    - Deploy broker on `cluster-a`

      Add `publicAPIServerURL` in `./config/samples/deploy_broker.yaml`, find the public apiserver URL with command: `kubectl config view  | grep server | cut -f 2- -d ":" | tr -d " "`

      ```shell
      kubectl -n knitnet-operator-system apply -f ./config/samples/deploy_broker.yaml
      ```

    - Export `submariner-broker-info` configmap to a yaml file

      ```shell
      kubectl -n knitnet-operator-system get cm submariner-broker-info -oyaml > submariner-broker-info.yaml
      ```

1. Join cluster to broker

     - Install knitnet operator

        Switch to `cluster-b`

        ```shell
        kubectl config use-context cluster-b
        ```

        Deploy operator

        ```shell
        make deploy
        ```

     - Create `submariner-broker-info` configmap

       ```shell
       kubectl create ns submariner-k8s-broker
       kubectl apply -f submariner-broker-info.yaml
       ```

     - Join `cluster-b` to `cluster-a`

       ```shell
       kubectl -n knitnet-operator-system apply -f ./config/samples/join_broker.yaml
       ```

### Verify

1. Deploy ClusterIP service on `cluster-b`

    Switch to `cluster-b`

    ```shell
    kubectl config use-context cluster-b
    ```

    Deploy `nginx` service

    ```shell
    kubectl -n default create deployment nginx --image=nginx
    kubectl -n default expose deployment nginx --port=80
    ```

1. Export service

   Create following resource on `cluster-b`:

    ```shell
    kubectl apply -f - <<EOF
    apiVersion: multicluster.x-k8s.io/v1alpha1
    kind: ServiceExport
    metadata:
      name: nginx
      namespace: default
    EOF
    ```

1. Run `nettest` from `cluster-a` to access the nginx service:

    Switch to `cluster-a`

    ```shell
    kubectl config use-context cluster-a
    ```

    Start `nettest` pod for test

    ```shell
    kubectl -n default  run --generator=run-pod/v1 tmp-shell --rm -i --tty --image quay.io/submariner/nettest -- /bin/bash
    ```

    Try to curl nginx service created in `cluster-b`

    ```shell
    curl nginx.default.svc.clusterset.local
    ```
