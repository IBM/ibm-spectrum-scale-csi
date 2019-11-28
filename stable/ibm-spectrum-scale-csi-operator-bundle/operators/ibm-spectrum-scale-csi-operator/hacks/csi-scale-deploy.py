#!/bin/python
# TODO write a script to break up yaml into multiple files.

import argparse
import sys
import os
import yaml
import socket
import subprocess
from change_deploy_image import change_image

BASE_DIR="{0}/..".format(os.path.dirname(os.path.realpath(__file__)))

def main(args):

  parser = argparse.ArgumentParser(
      description='''A hack to launch or tear down the operator.''')
  
  parser.add_argument( '-b', '--build',  dest='build', 
      action='store_true',
      help='''A flag to trigger the build.''')

  parser.add_argument( '-d', '--delete',  dest='delete', 
      action='store_true',
      help='''A flag to delete the operator and its CRDs.''')

  parser.add_argument( '-r','--repo',  metavar='repo', dest='repo', 
     default=socket.getfqdn(),
      help='''The target repository, defaults to fqdn of current node.''')

  args = parser.parse_args()

  if args.build:
    # Build the image 
    print BASE_DIR
    o = subprocess.Popen("/bin/bash -c 'operator-sdk build csi-scale-operator'", 
        env={"GO111MODULE": "on"}, 
        cwd=BASE_DIR, 
        shell=True,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = o.communicate()
    print("{0}\n{1}".format(out, err))

    # Update to target image 
    targetimage="{0}/csi-scale-operator".format(args.repo)
    change_image(image=targetimage)

    # tag and push to repository 
    o = subprocess.Popen("/bin/bash -c 'docker tag csi-scale-operator {0}'".format(targetimage), 
        shell=True,
        cwd=BASE_DIR, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = o.communicate()
    print("{0}\n{1}".format(out, err))

    o = subprocess.Popen("/bin/bash -c 'docker push {0}'".format(targetimage),
        shell=True,
        cwd=BASE_DIR, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = o.communicate()
    print("{0}\n{1}".format(out, err))

  return
  # Create the operator
  kaction="delete" if args.delete else "apply"
  kubeactions= [
    "/bin/bash -c 'kubectl {0} -f deploy/role.yaml",
    "/bin/bash -c 'kubectl {0} -f deploy/service_account.yaml",
    "/bin/bash -c 'kubectl {0} -f deploy/role_binding.yaml",
    "/bin/bash -c 'kubectl {0} -f deploy/crds/ibm-spectrum-scale-csi-operator-crd.yaml",
    "/bin/bash -c 'kubectl {0} -f deploy/operator.yaml" 
  ]

  for action in kubeactions:
    o = subprocess.Popen(action.format(kaction), 
        cwd=BASE_DIR, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = o.communicate()
    print("{0}\n{1}".format(out, err))

if __name__ == "__main__":
    sys.exit(main(sys.argv))

