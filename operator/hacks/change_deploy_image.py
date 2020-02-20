#!/bin/python
# TODO write a script to break up yaml into multiple files.

import argparse
import sys
import os
import yaml

DEFAULT_DEPLOY="{0}/../deploy/operator.yaml".format(os.path.dirname(os.path.realpath(__file__)))
DEFAULT_IMAGE="quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-operator:v1.1.0"
DEV_PULL="Always"
REL_PULL="IfNotPresent"
NEV_PULL="Never"

def change_image(operator=DEFAULT_DEPLOY, output=DEFAULT_DEPLOY, image=DEFAULT_IMAGE, pullpolicy=REL_PULL ):
  with open(operator, 'r') as stream:
    try:
        opobj= yaml.safe_load(stream)
    except yaml.YAMLError as e:
        print(e)
        return 1
    
    containers=opobj.get("spec",{}).get("template",{}).get("spec",{}).get("containers",[])

    for i, container in enumerate(containers):
      containers[i]["image"] = image
      containers[i]["imagePullPolicy"] = pullpolicy

    with open(output, 'w') as outfile:
      yaml.dump(opobj, outfile, default_flow_style=False)


def main(args):

  parser = argparse.ArgumentParser(
      description='''
A hack to change the image used by the operator deployment.
This should be used when deploying a custom image.
''')
  
  parser.add_argument( '-o', '--operator', metavar='operator', dest='operator', 
      default=DEFAULT_DEPLOY,
      help='''The filename of the operator deployment to edit.''')

  parser.add_argument( '-i','--image',  metavar='image', dest='image', 
      default=DEFAULT_IMAGE,
      help='''The new image, defaults to "official" image.''')

  parser.add_argument( '--ifnotpresent',  dest='ifnotpresent', 
      action='store_true',
      help='''A flag to set image pull to ifnotpresent.''')

  parser.add_argument( '--neverpull',  dest='neverpull', 
      action='store_true',
      help='''A flag to set never pull the image, takes precedence.''')


  parser.add_argument( '--output', metavar='output', dest='output',
     default=DEFAULT_DEPLOY,
     help='''Where to save the deployment, if unset overwrites deployment.''')

  args = parser.parse_args()

  if args.operator is None:
      print("Missing operator filename")
      return 1

  # Prep the operator input.
  operator=args.operator
  if operator[0] is not '/':
    operator="{0}/{1}".format(os.getcwd(), operator)
  
  # Prep the output
  output=args.output
  if output[0] is not '/':
      output ="{0}/{1}".format(os.getcwd(), output)

  image      = args.image
  pullpolicy = REL_PULL if (image == DEFAULT_IMAGE or args.ifnotpresent) else DEV_PULL

  if args.neverpull:
    pullpolicy = NEV_PULL

  change_image(operator, output, image, pullpolicy)


if __name__ == "__main__":
    sys.exit(main(sys.argv))

