#!/bin/python

import argparse
import sys
import os
import yaml
import subprocess
import textfsm

class Cluster:
  def __init__(self):
    self.id = ""
    self.name = ""
    self.nodes = []


def main(args):
    parser = argparse.ArgumentParser( 
        description='''A python script to scrape the output of mmlscluster for ingestion in ansible.''')
    parser.add_argument( '-c', metavar="command", dest='command', default="/usr/lpp/mmfs/bin/mmlscluster",
        help='''The command for mmlscluster, only specify if installed somewhere than the default.''')
    args = parser.parse_args()

    p = subprocess.Popen([args.command], stdout=subprocess.PIPE, 
                                         stderr=subprocess.PIPE)
    out, err = p.communicate()

    output = Cluster()
    with open("mmlscluster.template","r") as template:
      re_table = textfsm.TextFSM(template)
      data = re_table.ParseText(out)

      zipped = dict()
      for row in data:
        zipped = dict(zip(re_table.keys(), row))
        node = zipped.get("Node", None)
        if node is not None:
            output.nodes.append(node)
      
      output.id   = zipped.ID
      output.name = zipped.Name


if __name__ == "__main__":
    sys.exit(main(sys.argv))
