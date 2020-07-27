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
Use the olm test playbook. 
First setup a [quay application repo](https://ibm-spectrum-scale-csi.readthedocs.io/en/latest/developers/olm.html)
Then customize `olm-test-playbook.yaml` to your environment (there's inline docs, _don't_ touch the Jinja2 templates).

```
  ansible-playbook olm-test-playbook.yaml
```
