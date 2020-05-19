# cert-vault
* A controller use vault go sdk generate cert. You need configure vault service address and root token as environment parameters in deploymnet.

* The crd looks like:
```
apiVersion: cert.vault.com/v1
kind: CertInfo
metadata:
  name: certinfo-sample
spec:
  role: test                       // role for generate cert
  ou: k8s
  max_ttl: 12h
  common_name: zh
  path: pki_demo                   // root ca save path
```

* The response looks like:
```
apiVersion: v1
data:
  certificate: LS0tLS1CRUdJTiBDRVJUSU
  issuing_ca: LS0tLS1CRUdJTi
  path: cGtpX2RlbW8=
  private_key: LS0tLS1CRUdJT
  serial_number: MGY6NTg6YzE6OGM6M
kind: Secret
metadata:
  creationTimestamp: "2020-05-19T02:06:10Z"
  name: certinfo-sample
  namespace: test
  ownerReferences:
  - apiVersion: cert.vault.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: CertInfo
    name: certinfo-sample
    uid: 3445a9b1-f16a-4bc5-865c-816cbf696ef4
  resourceVersion: "67479472"
  selfLink: /api/v1/namespaces/test/secrets/certinfo-sample
  uid: 759ebfd9-f1d1-4fb7-aea2-91392fcedcc7
type: Opaque
```

