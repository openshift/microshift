# Known Limitations

Despite the list of enabled APIs is documented, there might be some cases where it's not clear whether a certain functionality works. This is a list of known limitations within MicroShift:

## Idling/Unidling Applications

The `oc` client provides a way to idle applications in order to reduce power consumption. By adding specific annotations, `oc` will scale all workloads referenced to a service down to zero. This is a fully working feature in MicroShift.

However, when the application service/route receives network traffic, OVN-Kubernetes will detect it and send an event to scale that application back up. MicroShift does not provide the unidling controller though, so this functionality won't work.
