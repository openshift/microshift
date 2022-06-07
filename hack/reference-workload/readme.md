# Reference Workload

This directory contains the manifects and necessary scripts to deploy
a reference application into MicroShift.

The architecture of this reference application looks like:

```
    amq-broker        app-logic           app-web              app-plugin1

           ┌────────────────────────────────────┐
           │                                    │
    ┌──────┴─────┐     ┌───────────┐     ┌──────┴───────┐
    │            │     │           │     │              │
    │ amq-broker │◄────┤ app-core  │ ◄───┤ app-web      │
    │            │     │           │     │              │
    └────────────┘     ├───────────┤     └───┬──────────┘
        ▲ ▲  ▲  ▲      │           │         │         ▲
        │ │  │  └──────┤ app-api   │◄────────┘         │      ┌─────────────┐
        │ │  │         │           │                   │      │             │
        │ │  │         └───────────┘◄──────────────────┼──────┤ app-plugin1 │
        │ │  │                            (only this)  │      │             │
        │ │  │                                         │      └─────────────┘
        │ │  │                                         │
 ────── │ │  │  ───────────────────────────────────────┼─────────────────────
        │ │  │                                         │            LAN
             │                                         │
AMQP & MQTT  │                                         │
(NodePorts)  amq-microshift.local              app-interface.local
             (Route)                           (Route)

```

Can be applied via:
`kubectl apply -k .`

While this application doesn´t do anything meaningful (yet), it serves as
a reference in terms of features and workloads to be ran.

The purpose of this reference workload is to help measure the impact on
MicroShift, the CNI (and eventually CSI) resource consumptions.