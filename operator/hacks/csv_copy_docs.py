#!/bin/python

import argparse
import sys
import os
import yaml

BASE_DIR="{0}/../".format(os.path.dirname(os.path.realpath(__file__)))
DEFAULT_VERSION="2.0.0"
CSV_PATH="{0}deploy/olm-catalog/ibm-spectrum-scale-csi-operator/{1}/ibm-spectrum-scale-csi-operator.v{1}.clusterserviceversion.yaml"
QUICKSTART="{0}../../../../docs/source/get-started/quickstart.md".format(BASE_DIR)

def main(args):
  parser = argparse.ArgumentParser(
    description='''A hack to copy docs into the CSV.''')
  
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
    spec = csv.get("spec",{})  

    # Set the description of the CSV to the quickstart. 
    with open(QUICKSTART, 'r') as stream:
     spec["description"] = stream.read() 
  

  with open(csvf, 'w') as outfile:
    yaml.dump(csv, outfile, default_flow_style=False)
      
if __name__ == "__main__":
    sys.exit(main(sys.argv))

