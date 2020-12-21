# Ansible overview

## Bundling operator for release

```
  ansible-playbook cert-playbook.yaml
```
Outputs to `../../build`.

 * `ibm-spectrum-scale-csi.zip` to be uploaded to redhat.
 * `ibm-spectrum-scale-csi-operator.zip` needs to be unzipped and placed in the upstream and community respectively
   * (upstream)[https://github.com/operator-framework/community-operators/pull/1871]
   * (community)[https://github.com/operator-framework/community-operators/pull/1870]

## Changed a CRD or file in operator/deploy?
Run the generate playbook:

```
  ansible-playbook generate-playbook.yaml
```

## Bumping a version?
Run the versioning playbook after updating `VERSION_NEW` and `VERSION_OLD` respectively in the playbook.

```
  ansible-playbook versioning-playbook.yaml
```

## Testing OLM? 

Follow the instructon on page [olmupgrade.md](./tools/ansible/olmupgrade.md)
