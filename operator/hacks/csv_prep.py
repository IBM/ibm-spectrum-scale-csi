#!/bin/python

import argparse
import sys
import os
import yaml

BASE_DIR="{0}/../".format(os.path.dirname(os.path.realpath(__file__)))
DEFAULT_VERSION="2.0.0"
CSV_PATH="{0}deploy/olm-catalog/ibm-spectrum-scale-csi-operator/{1}/ibm-spectrum-scale-csi-operator.v{1}.clusterserviceversion.yaml"

def main(args):
  parser = argparse.ArgumentParser(
    description='''A hack to prep the CSV for regeneration.''')
  
  parser.add_argument( '--version', metavar='CSV Version', dest='version', default=DEFAULT_VERSION,
    help='''The version of the CSV to update''')

  args = parser.parse_args()
  
  csvf = CSV_PATH.format(BASE_DIR, args.version)
  csv = None
  try:
    with open(csvf, 'r') as stream:
      csv = yaml.safe_load(stream)
  except yaml.YAMLError as e:
    print(e)
    return 1

  # Edit the contents of the CSV 
  if csv is not None:
    csv.get("spec",{}).pop("install", None)


  with open(csvf, 'w') as outfile:
    yaml.dump(csv, outfile, default_flow_style=False)
      
if __name__ == "__main__":
    sys.exit(main(sys.argv))

