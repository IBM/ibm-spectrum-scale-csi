#!/bin/python
# TODO write a script to break up yaml into multiple files.

import argparse
import sys
import os
import yaml

BASE_DIR="{0}/..".format(os.path.dirname(os.path.realpath(__file__)))
DEPLOY_DIR="{0}/deploy".format(BASE_DIR)
CONFIG_FILES=["operator.yaml", "role_binding.yaml", "role.yaml", "service_account.yaml",
    "crds/ibm_v1alpha1_csiscaleoperator_crd.yaml", "crds/ibm_v1alpha1_csiscaleoperator_cr.yaml"]


def label(obj):
  if "metadata" not in obj.keys():
    obj["metadata"] = {"labels":{}}
  elif "labels" not in obj["metadata"].keys():
    obj["metadata"]["labels"] = {}

  metadata = obj["metadata"]
  labels   = metadata["labels"]
  name     = metadata.get("name","").lower()
  kind     = obj.get("kind", "").lower()

  labels["app.kubernetes.io/instance"]   = "ssco-{0}".format(name)
  labels["app.kubernetes.io/managed-by"] = "operator"
  labels["app.kubernetes.io/name"]       = "spectrum-scale-csi-operator"


def main(args):
    parser = argparse.ArgumentParser(
        description='''A hack apply the default labels to the resources.''')
    
    args = parser.parse_args()

    # Load the exiting files and update the labels. 
    for config in CONFIG_FILES:
      filepath="{0}/{1}".format(DEPLOY_DIR, config)

      # Open the old file and get a list of objects. 
      objs = []
      with open(filepath, 'r') as stream:
        try:
            objs = list(yaml.safe_load_all(stream))
        except yaml.YAMLError as e:
            print(e)
            return 1

      # Relabel
      for obj in objs:
        if obj is None:
          continue
        label(obj)

      # Update with fresh labels.
      #os.remove(filepath)
      with open(filepath, 'w') as outfile:
        yaml.dump_all(objs, outfile, default_flow_style=False)


if __name__ == "__main__":
    sys.exit(main(sys.argv))

