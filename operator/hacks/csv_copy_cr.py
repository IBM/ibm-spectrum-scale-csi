#!/bin/python

import argparse
import sys
import os
import yaml
import json

BASE_DIR="{0}/../".format(os.path.dirname(os.path.realpath(__file__)))
DEFAULT_VERSION="2.0.0"
CSV_PATH="{0}deploy/olm-catalog/ibm-spectrum-scale-csi-operator/{1}/ibm-spectrum-scale-csi-operator.v{1}.clusterserviceversion.yaml"
CR="{0}/deploy/crds/{1}"

def main(args):
  parser = argparse.ArgumentParser(
    description='''A hack to copy commented CRS into the CSV.''')
  
  parser.add_argument( '--cr', metavar='cr', dest='cr', default=None,
      help='''The Custom Resource File.''')

  parser.add_argument( '--version', metavar='CSV Version', dest='version', default=DEFAULT_VERSION,
    help='''The version of the CSV to update''')


  args = parser.parse_args()

  crf=CR.format(BASE_DIR,args.cr)
  csvf = CSV_PATH.format(BASE_DIR, args.version)

  csv = None
  cr = None
  try:
    with open(crf, 'r') as stream:
      cr = yaml.safe_load(stream)
    with open(csvf, 'r') as stream:
      csv = yaml.safe_load(stream)
  except yaml.YAMLError as e:
    print(e)
    return 1

  # Remove namespace from the CR and update CSV
  if cr is not None and csv is not None:
    annotations = csv.get("metadata",{}).get("annotations",{})
    cr.get("metadata",{}).pop("namespace", None)
    annotations["alm-examples"] = json.dumps(cr)

    with open(csvf, 'w') as outfile:
      yaml.dump(csv, outfile, default_flow_style=False)
      
if __name__ == "__main__":
    sys.exit(main(sys.argv))

