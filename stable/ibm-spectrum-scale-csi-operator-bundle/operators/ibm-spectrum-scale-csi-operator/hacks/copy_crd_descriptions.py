#!/bin/python
# TODO write a script to break up yaml into multiple files.

import argparse
import sys
import os
import yaml

def loadDescriptors(descType, crd, csv):
    props = crd.get("spec", {})    \
        .get("validation",{})      \
        .get("openAPIV3Schema", {})\
        .get("properties", {})     \
    
    spec            = props.get(descType, {})
    specprops       = spec.get("properties", {})
    specdescriptors = []
    specpaths       = specprops.keys()
    
    while len(specpaths) > 0:
        # Load the path and then iterate to get the property.
        path    = specpaths.pop(0)
        keys    = path.split(".")
        prop    = spec
        subprops = specprops
      
        for key in keys:
            prop     = subprops.get(key, {})
            subprops = prop.get("properties", {})
        
        # Enqueue the sub properties (if present)
        for subprop in prop.get("properties", {}).keys():
            specpaths.append("{0}.{1}".format(path,subprop))
        
        # Construct description
        specdescriptors.append({ 
          "path" : path,
          "description" : prop.get("description", "") })
        
    return specdescriptors



def main(args):
    parser = argparse.ArgumentParser(
        description='''A hack to clone descriptions from the CRD.''')
    
    parser.add_argument( '--crd', metavar='crd', dest='crd', default=None,
        help='''The Custom Resource Definition File.''')

    parser.add_argument( '--csv', metavar='csv', dest='csv', default=None,
        help='''The output CSV to clone the data to.''')

    args = parser.parse_args()

    if args.crd is None:
        print("Missing Custom Resource Definition")
        return 1
    if args.csv is None:
        print("Missing CSV")
        return 1
    
    crdf=args.crd
    if crdf[0] is not '/':
        crdf ="{0}/{1}".format(os.getcwd(), crdf)

    csvf=args.csv
    if csvf[0] is not '/':
        csvf ="{0}/{1}".format(os.getcwd(), csvf)

    crd = None
    csv = None
    try:
        with open(crdf, 'r') as stream:
            crd = yaml.safe_load(stream)
        with open(csvf, 'r') as stream:
            csv = yaml.safe_load(stream)
    except yaml.YAMLError as e:
        print(e)
        return 1

    if crd is not None and csv is not None:
        metaname= crd.get("metadata",{}).get("name", "")
        specdescriptors = loadDescriptors("spec", crd, csv)
        statusdescriptors = loadDescriptors("status", crd, csv)

        # TODO need to replace with something 
        resourcefound=False
        owned = csv.get("spec",{}).get("customresourcedefinitions",{}).get("owned",{})
        for resource in owned:
            if resource.get("name","") == metaname:
                resourcefound = True
                resource["specDescriptors"] = specdescriptors
                resource["statusDescriptors"] = statusdescriptors
                resource["version"] = crd.get("spec",{}).get("version","v1")

        # If the resource wasn't present in the customresourcedefinitions add it. 
        if not resourcefound:
            owned.append({
                "name"            : metaname,
                "kind"            : crd.get("kind","CustomResourceDefinition"),
                "version"         : crd.get("spec",{}).get("version","v1"),
                "displayName"     : metaname,
                "specDescriptors" : specdescriptors,
                "description"     : "TODO: Fill this in"})


        with open(csvf, 'w') as outfile:
            yaml.dump(csv, outfile, default_flow_style=False)
        




if __name__ == "__main__":
    sys.exit(main(sys.argv))

