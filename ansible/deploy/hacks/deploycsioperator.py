#!/bin/python

import argparse
import sys
import os
import yaml
import subprocess
import textfsm


def mmlscluster(cmd, tmpl):
  ''' Run the mmlscluster command and grab the relevant data '''

  p = subprocess.Popen([cmd], stdout=subprocess.PIPE, 
                              stderr=subprocess.PIPE)
  out, err = p.communicate()

  output = {
      "id": None,
      "name": None,
      "nodes":[] 
  }
  with open(tmpl,"r") as template:
    re_table = textfsm.TextFSM(template)
    data = re_table.ParseText(out)

    zipped = dict()
    for row in data:
      zipped = dict(zip(re_table.header, row))
      node = zipped.get("Node", None)
      if node is not None and isinstance(node, str):
          output["nodes"].append(node)
    
    output["id"]   = zipped.get("ID")
    output["name"] = zipped.get("Name")

    return output

def mmlsgui(cmd, tmpl):
  ''' Run mmlsnodeclass GUI_MGMT_SERVERS, return the first instance found. '''
   
  p = subprocess.Popen([cmd, "GUI_MGMT_SERVERS"], stdout=subprocess.PIPE,
                                                  stderr=subprocess.PIPE)
  out, err = p.communicate()
  
  with open(tmpl,"r") as template:
    re_table = textfsm.TextFSM(template)
    data = re_table.ParseText(out)
      
    zipped = dict()
    for row in data:
      zipped = dict(zip(re_table.header, row))
      gui = zipped.get("GUI_MGMT_SERVER", None)
      if gui is not None and isinstance(gui, str):
        return gui

  return ""

def mmlsfs(cmd, tmpl):
  ''' Run the mmlsfs command and grab the relevant data. '''
  p = subprocess.Popen([cmd, "all", "-T"], 
    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
  out, err = p.communicate()

  output=[]
  with open(tmpl,"r") as template:
    re_table = textfsm.TextFSM(template)
    data = re_table.ParseText(out)
      
    zipped = dict()
    for row in data:
      zipped = dict(zip(re_table.header, row))

      output.append({
          "fs"    :  zipped.get("fs", ""),
          "mount" :  zipped.get("mount", "")
          })

  return output 


def main(args):
  parser = argparse.ArgumentParser( 
      description='''A python script to scrape the output of mmlscluster for ingestion in ansible.''')
  parser.add_argument( '--mmlscluster', metavar="mmlscluster", 
      dest='mmlscluster', default="/usr/lpp/mmfs/bin/mmlscluster",
      help='''The command for mmlscluster, only specify if installed somewhere than the default.''')
  parser.add_argument( '--mmlsnodeclass', metavar="mmlsnodeclass", 
      dest='mmlsnodeclass', default="/usr/lpp/mmfs/bin/mmlsnodeclass",
      help='''The command for mmlsnodeclass, only specify if installed somewhere than the default.''')
  parser.add_argument( '--mmlsfs', metavar="mmlsfs", 
      dest='mmlsfs', default="/usr/lpp/mmfs/bin/mmlsfs",
      help='''The command for mmlsfs, only specify if installed somewhere than the default.''')
  parser.add_argument( '--templates', metavar='templates',
      dest='templates', default=os.path.dirname(os.path.realpath(__file__)),
      help='''The directory containing the TextFSM templates''')


  parser.add_argument('--fs', metavar='filesystem',
      dest='fs', default=None,
      help='''The name of the filesystem (including /dev/) to use with the csi driver'''

  parser.add_argument('--fset', metavar='fileset',
      dest='fs', default=None,
      help='''The name of the filesystem (including /dev/) to use with the csi driver'''

  args = parser.parse_args()

  # Create the template paths
  SCRIPT_DIR = args.templates
  MMLSCLUSTER_TEMPL = "{0}/templates/mmlscluster".format(SCRIPT_DIR)
  MMLSNODECLASS_TEMPL = "{0}/templates/mmlsnodeclass".format(SCRIPT_DIR)
  MMLSFS_TEMPL = "{0}/templates/mmlsfs".format(SCRIPT_DIR)

  cluster=mmlscluster(args.mmlscluster,        MMLSCLUSTER_TEMPL)
  cluster["gui"] = mmlsgui(args.mmlsnodeclass, MMLSNODECLASS_TEMPL)
  cluster["fs"]  = mmlsfs(args.mmlsfs,         MMLSFS_TEMPL)

  print(yaml.dump(cluster))

if __name__ == "__main__":
    sys.exit(main(sys.argv))
