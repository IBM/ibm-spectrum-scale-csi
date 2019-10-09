#!/bin/python
# TODO write a script to break up yaml into multiple files.

import argparse
import sys
import os
import yaml

def main(args):
    parser = argparse.ArgumentParser(
        description='''A hack to split the YAML objects in a file into YAML multiple files.''')
    
    parser.add_argument( '-f', metavar='filename', dest='filename', default=None,
        help='''The filename containing multiple YAML objects.''')
    parser.add_argument( '-d', metavar='destination', dest='dest', default='.',
        help='''The destination to store the contents of the YAML files.''')
    parser.add_argument( '-p', metavar='prefix', dest='prefix', default='ibm-csi-scale',
        help='''The prefix for the generated file names.''')

    args = parser.parse_args()

    if args.filename is None:
        print("Missing filename")
        return 1
    
    dest=args.dest
    if dest[0] is not '/':
        dest ="{0}/{1}".format(os.getcwd(), dest)

    with open(args.filename, 'r') as stream:
        try:
            objs = yaml.safe_load_all(stream)
        except yaml.YAMLError as e:
            print(e)
            return 1
        

        for  obj in objs:
            if obj is None:
                continue

            ofile="{0}/{1}-{2}_{3}.yaml".format(dest, args.prefix, obj.get("kind", ""), 
                obj.get("metadata",{}).get("name", ""))

            os.remove(ofile)
            with open(ofile, 'w') as outfile:
                yaml.dump(obj, outfile, default_flow_style=False)


if __name__ == "__main__":
    sys.exit(main(sys.argv))

