#!/bin/python

import argparse
import sys
import os
import yaml
import shutil

BASE_DIR="{0}/../".format(os.path.dirname(os.path.realpath(__file__)))
WATCHES= '{0}watches.yaml'.format(BASE_DIR)
BACKUP = '{0}watches.yaml.bak'.format(BASE_DIR)

def main(args):
  parser = argparse.ArgumentParser(
    description='''A hack to strip and restore finalizers.''')

  parser.add_argument( '-r', '--restore',  dest='restore', 
      action='store_true',
      help='''A flag to restore finalizers.''')

  args = parser.parse_args()

  # If restore is set apply the backup
  if args.restore: 
    if os.path.isfile(BACKUP):
      shutil.copy(BACKUP, WATCHES)
      os.remove(BACKUP)

  # Else  clear the finalizers
  else:
    watches = []
    try:
      with open(WATCHES, 'r') as stream:
        watches = yaml.safe_load(stream)
    except yaml.YAMLError as e:
      print(e)
      return 1
    
    # Pop the finalizer 
    for watch in watches:
      watch.pop('finalizer', None)
    
    shutil.copy(WATCHES, BACKUP)
    # Backup and dump the new watches.
    with open(WATCHES, 'w') as outfile:
      yaml.dump(watches, outfile, default_flow_style=False)



      
if __name__ == "__main__":
    sys.exit(main(sys.argv))

