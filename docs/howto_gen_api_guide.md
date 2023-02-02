# Generating the REST API Guide

The MicroShift REST API supports a subset of the OCP APIs and some
additional APIs not present in OCP by default. We therefore generate
the files for the API guide from a MicroShift deployment, instead of
manually curating the set of API definitions from the OCP
documentation.

1. Set up a host with MicroShift following any of the sets of
   instructions for creating a development environment.

2. Ensure `podman` is installed.

   $ sudo dnf -y install podman

3. Ensure the `microshift` user has a valid `~/.kube/config` file for
   the local deployment.

   $ mkdir -p ~/.kube
   $ sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > .kube/config

4. From the root of the local `microshift` git repository, run the
   script to generate the asciidoc files.

   $ ./hack/gen-api-docs.sh

The asciidoc files are written to
`./_output/openshift-apidocs-gen/microshift_rest_api`. Copy that
directory and its contents to replace the same directory in the
`openshift-docs` git repository.

The topic map content is written to
`./_output/openshift-apidocs-gen/_topic_map_segment.yml`. Update the
"API Reference" section of the file `_topic_maps/_topic_map_ms.yml` in
the `openshift-docs` repository using the contents of the newly
generated file. Make sure to retain the comments before and after the
"API Reference" section.
