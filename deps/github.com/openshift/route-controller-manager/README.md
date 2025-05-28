# Route Controller Manager

The Route Controller Manager consists of additional controllers that enhance Openshift Routes, Ingresses, and Services.

## Ingress to Route Controller

Controller ensures that zero or more routes exist to match any supported ingress. The
controller creates a controller owner reference from the route to the parent ingress,
allowing users to orphan their ingress. All owned routes have specific spec fields
managed (those attributes present on the ingress), while any other fields may be
modified by the user.


## Service Ingress IP Controller

Controller is responsible for allocating ingress ip addresses to Service objects of type LoadBalancer. 
It allocates adresses from `spec.observedConfig.ingress.ingressIPNetworkCIDR` range in
`openshiftcontrollermanagers.operator.openshift.io cluster` config and can be used to assign a unique external IP addresses.

