You usually want to start with NewSimpleClientCertificateController.

This package provides a control loop which takes as input
1. target secret name
2. cert common name
3. desired validity (recall that the signing cert can sign for less)

The flow goes like this.
1. if secret contains a valid client cert good for at least five days or 50% of validity, do nothing.  If not...
2. create new cert/key pair in memory
3. create CSR in the API.
4. watch CSR in the API until it is approved or denied
5. if denied, write degraded status and return
6. if approved, update the secret

The secrets have annotations which match our other cert rotation secrets.